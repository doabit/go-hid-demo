
package main

import (
	"time"

	"github.com/karalabe/hid"
)

// UsbDevice represents a USB device
type UsbDevice struct {
	Path       string
	VendorID   uint16
	ProductID  uint16
}

// Event is sent from the Observer
type Event int

const (
	// Initial list of devices when polling starts
	Initial Event = iota
	// A device that has just been connected
	Connect
	// A device that has just disconnected
	Disconnect
)

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
	vendorID     *uint16
	productID    *uint16
}

// NewObserver creates a new Observer with the given poll interval
func NewObserver(pollInterval time.Duration) *Observer {
	return &Observer{
		pollInterval: pollInterval,
	}
}

// WithVendorID filters results by USB Vendor ID
func (o *Observer) WithVendorID(vendorID uint16) *Observer {
	o.vendorID = &vendorID
	return o
}

// WithProductID filters results by USB Product ID
func (o *Observer) WithProductID(productID uint16) *Observer {
	o.productID = &productID
	return o
}

// Subscribe starts the background thread and polls for device changes
func (o *Observer) Subscribe() *Subscription {
	rxEvent := make(chan Event)
	txClose := make(chan struct{})

	go func() {
		api := hid.NewHidApi()

		deviceList, err := api.Enumerate(uint16(0), uint16(0))
		if err != nil {
			close(rxEvent)
			return
		}

		var usbDevices []UsbDevice
		for _, deviceInfo := range deviceList {
			if (o.vendorID == nil || *o.vendorID == deviceInfo.VendorId) &&
				(o.productID == nil || *o.productID == deviceInfo.ProductId) {
				usbDevices = append(usbDevices, UsbDevice{
					Path:       deviceInfo.Path,
					VendorID:   deviceInfo.VendorId,
					ProductID:  deviceInfo.ProductId,
				})
			}
		}

		rxEvent <- Initial

		ticker := time.NewTicker(o.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				nextDeviceList, err := api.Enumerate(uint16(0), uint16(0))
				if err != nil {
					close(rxEvent)
					return
				}

				var nextUsbDevices []UsbDevice
				for _, deviceInfo := range nextDeviceList {
					if (o.vendorID == nil || *o.vendorID == deviceInfo.VendorId) &&
						(o.productID == nil || *o.productID == deviceInfo.ProductId) {
						nextUsbDevices = append(nextUsbDevices, UsbDevice{
							Path:       deviceInfo.Path,
							VendorID:   deviceInfo.VendorId,
							ProductID:  deviceInfo.ProductId,
						})
					}
				}

				for _, device := range usbDevices {
					if !contains(nextUsbDevices, device) {
						rxEvent <- Disconnect
					}
				}

				for _, device := range nextUsbDevices {
					if !contains(usbDevices, device) {
						rxEvent <- Connect
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
	observer := NewObserver(time.Second).
		WithVendorID(0x1234).
		WithProductID(0x5678)

	subscription := observer.Subscribe()

	defer close(subscription.TxClose)

	for event := range subscription.RxEvent {
		switch event {
		case Initial:
			// handle initial list of devices
		case Connect:
		case Disconnect:

		}
	}
