以下是使用golang实现的代码，需要先安装hid库：go get github.com/karalabe/hid/v2

package main

import (
    "fmt"
    "github.com/karalabe/hid/v2"
    "sync"
    "time"
)

// UsbDevice represents a USB device
type UsbDevice struct {
    // Path
    Path string
    // Vendor ID
    VendorId uint16
    // Product ID
    ProductId uint16
}

// Event represents the events sent from the observer
type Event int

const (
    // Initial event, contains the list of devices when polling starts
    InitialEvent Event = iota
    // A device has just been connected
    ConnectEvent
    // A device has just been disconnected
    DisconnectEvent
)

// Observer observes USB device changes
type Observer struct {
    pollInterval time.Duration
    vendorId     *uint16
    productId    *uint16
}

// Subscription is used to subscribe to events from the observer
type Subscription struct {
    RxEvent <-chan Event
    close   chan<- struct{}
}

// NewObserver creates a new observer instance with default configuration
func NewObserver() *Observer {
    return &Observer{
        pollInterval: 1 * time.Second,
        vendorId:     nil,
        productId:    nil,
    }
}

// WithPollInterval sets the polling interval
func (o *Observer) WithPollInterval(interval time.Duration) *Observer {
    o.pollInterval = interval
    return o
}

// WithVendorID filters the results by USB vendor ID
func (o *Observer) WithVendorID(vendorID uint16) *Observer {
    o.vendorId = &vendorID
    return o
}

// WithProductID filters the results by USB product ID
func (o *Observer) WithProductID(productID uint16) *Observer {
    o.productId = &productID
    return o
}

// Subscribe subscribes to the events from the observer
func (o *Observer) Subscribe() *Subscription {
    rxEvent := make(chan Event)
    closeSignal := make(chan struct{}, 1)

    go func() {
        var deviceList []UsbDevice

        // Initialize HIDAPI
        if err := hid.Init(); err != nil {
            return
        }
        defer hid.Exit()

        // Get initial device list
        devices, err := hid.Enumerate(0, 0)
        if err == nil {
            for _, device := range devices {
                if (o.vendorId == nil || *o.vendorId == device.VendorID) &&
                    (o.productId == nil || *o.productId == device.ProductID) {

                    deviceList = append(deviceList, UsbDevice{
                        Path:      device.Path,
                        VendorId:  device.VendorID,
                        ProductId: device.ProductID,
                    })
                }
            }

            // Send initial event
            rxEvent <- InitialEvent
            rxEvent <- deviceList
        }

        // Polling loop
        waitSeconds := float32(o.pollInterval.Seconds())
        for {
            select {
            case <-closeSignal:
                return
            default:
                time.Sleep(time.Millisecond * 250)

                waitSeconds -= 0.25
                if waitSeconds <= 0 {
                    waitSeconds = float32(o.pollInterval.Seconds())

                    // Refresh device list
                    hid.RefreshDevices()

                    // Get current device list
                    devices, err := hid.Enumerate(0, 0)
                    if err != nil {
                        continue
                    }

                    nextDeviceList := make([]UsbDevice, 0)
                    for _, device := range devices {
                        if (o.vendorId == nil || *o.vendorId == device.VendorID) &&
                            (o.productId == nil || *o.productId == device.ProductID) {

                            nextDeviceList = append(nextDeviceList, UsbDevice{
                                Path:      device.Path,
                                VendorId:  device.VendorID,
                                ProductId: device.ProductID,
                            })
                        }
                    }

                    // Send events
                    for _, device := range deviceList {
                        if !containsDevice(nextDeviceList, device) {
                            rxEvent <- DisconnectEvent
                            rxEvent <- device
                        }
                    }

                    for _, device := range nextDeviceList {
                        if !containsDevice(deviceList, device) {
                            rxEvent <- ConnectEvent
                            rxEvent <- device
                        }
                    }

                    deviceList = nextDeviceList
                }
            }
        }
    }()

    return &Subscription{
        RxEvent: rxEvent,
        close:   closeSignal,
    }
}

// containsDevice checks whether the given device list contains a specific device
func containsDevice(devices []UsbDevice, device UsbDevice) bool {
   }
