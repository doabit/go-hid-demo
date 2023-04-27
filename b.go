package main

import (
	"fmt"
	"time"

	// hid "github.com/GeertJohan/go.hid"
	hid "github.com/karalabe/usb"
)

// UsbDevice represents a USB device
type UsbDevice struct {
	Path      string
	VendorID  uint16
	ProductID uint16
}

// Event is sent from the Observer
//
//	type Event  {
//		Device UsbDevice
//	}
type EventType int

const (
	// Initial list of devices when polling starts
	Initial EventType = iota
	// A device that has just been connected
	Connect
	// A device that has just disconnected
	Disconnect
)

type Event struct {
	Device UsbDevice
	Type   EventType
}

// Subscription is used to receive events from the Observer
type Subscription struct {
	RxEvent chan Event
	// When this gets closed, the channel will become disconnected and the
	// background loop will terminate
	TxClose chan struct{}
}

// Observer observes USB devices and sends events to subscribers
type Observer struct {
	pollInterval time.Duration
	vendorID     uint16
	productID    uint16
}

// NewObserver creates a new Observer with the given poll interval
func NewObserver(vendorID uint16, productID uint16, pollInterval time.Duration) *Observer {
	return &Observer{
		vendorID:     vendorID,
		productID:    productID,
		pollInterval: pollInterval,
	}
}

// WithVendorID filters results by USB Vendor ID
func (o *Observer) WithVendorID(vendorID uint16) *Observer {
	o.vendorID = vendorID
	return o
}

// WithProductID filters results by USB Product ID
func (o *Observer) WithProductID(productID uint16) *Observer {
	o.productID = productID
	return o
}

// Subscribe starts the background thread and polls for device changes
func (o *Observer) Subscribe() *Subscription {
	rxEvent := make(chan Event)
	txClose := make(chan struct{})

	go func() {
		// api := hid.NewHidApi()

		deviceList, err := hid.EnumerateHid(0x2020, 0x2020)
		if err != nil {
			close(rxEvent)
			return
		}

		var usbDevices []UsbDevice
		for _, deviceInfo := range deviceList {
			usbDevices = append(usbDevices, UsbDevice{
				Path:      deviceInfo.Path,
				VendorID:  deviceInfo.VendorID,
				ProductID: deviceInfo.ProductID,
			})
		}

		rxEvent <- Event{Device: UsbDevice{}, Type: Initial}

		ticker := time.NewTicker(o.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				nextDeviceList, err := hid.EnumerateHid(o.vendorID, o.productID)
				if err != nil {
					close(rxEvent)
					return
				}

				var nextUsbDevices []UsbDevice
				for _, deviceInfo := range nextDeviceList {
					nextUsbDevices = append(nextUsbDevices, UsbDevice{
						Path:      deviceInfo.Path,
						VendorID:  deviceInfo.VendorID,
						ProductID: deviceInfo.ProductID,
					})
				}

				for _, device := range usbDevices {
					if !contains(nextUsbDevices, device) {
						rxEvent <- Event{Device: device, Type: Disconnect}
					}
				}

				for _, device := range nextUsbDevices {
					if !contains(usbDevices, device) {
						rxEvent <- Event{Device: device, Type: Connect}
					}
				}

				usbDevices = nextUsbDevices

			case <-txClose:
				close(rxEvent)
				return
			}
		}
	}()

	return &Subscription{
		RxEvent: rxEvent,
		TxClose: txClose,
	}
}

func contains(devices []UsbDevice, device UsbDevice) bool {
	for _, d := range devices {
		if d.Path == device.Path &&
			d.VendorID == device.VendorID &&
			d.ProductID == device.ProductID {
			return true
		}
	}
	return false
}

func main() {
	observer := NewObserver(0x2020, 0x2020, time.Second)

	subscription := observer.Subscribe()

	defer close(subscription.TxClose)

	for event := range subscription.RxEvent {
		switch event.Type {
		case Initial:
			fmt.Println("intial")
			break
			// handle initial list of devices
		case Connect:
			fmt.Println("Connect", event.Device)
			break
		case Disconnect:
			fmt.Println("Disconnect", event.Device)
		}
	}
}
