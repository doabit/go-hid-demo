package main

import (
	"fmt"
	"time"
)

func main() {
	observer := NewObserver(0x2020, 0x2020, time.Second)

	subscription := observer.Subscribe()

	defer close(subscription.TxClose)

	for event := range subscription.RxEvent {
		switch event.Type {
		case Initial:
			fmt.Println("intial")
		case Connect:
			fmt.Println("Connect", event.Device)
		case Disconnect:
			fmt.Println("Disconnect", event.Device)
		}
	}
}
