package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients = make(map[*websocket.Conn]bool)
)

type Msg struct {
	Data string `json:"data"`
}

func mainTest() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	go sendMsg()
	http.ListenAndServe(":8080", nil)
}

func sendMsg() {
	for {
		time.Sleep(1 * time.Second)
		fmt.Println("xiao")
		msg := Msg{Data: "xiaojin"}
		msgJson, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("ok")
			return
		}
		broadcast(msgJson)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, conn)
			break
		}
		broadcast(msg)
	}
}

func broadcast(msg []byte) {
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println(err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
