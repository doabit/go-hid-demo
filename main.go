package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"wscast_demo/ws"
)

func usb() {
	observer := NewObserver(0x2020, 0x2020, time.Second)
	fmt.Println("usb")
	subscription := observer.Subscribe()

	defer close(subscription.TxClose)

	for event := range subscription.RxEvent {
		switch event.Type {
		case Initial:
			// ws.SendJsonMsg("intial", "pop")
			fmt.Println("intial")
		case Connect:
			ws.SendJsonMsg("Connect", "popUp")
			fmt.Println("Connect", event.Device)
		case Disconnect:
			ws.SendJsonMsg("Disconnect", "popOut")
			fmt.Println("Disconnect", event.Device)
		}
	}
}

func main() {
	go usb()

	fileServer := http.FileServer(http.Dir("./static/"))

	// Use the mux.Handle() function to register the file server as the handler for
	// all URL paths that start with "/static/". For matching paths, we strip the
	// "/static" prefix before the request reaches the file server.
	http.Handle("/static/", http.StripPrefix("/static", fileServer))

	http.HandleFunc("/ws", ws.HandleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	// go ws.SendMsg()
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
