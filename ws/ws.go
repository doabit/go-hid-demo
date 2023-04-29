package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
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
	Data   string `json:"data"`
	Action string `json:"type"`
}

func SendJsonMsg(msg string, action string) {
	jsonMsg, err := json.Marshal(Msg{Data: msg, Action: action})
	if err != nil {
		log.Println(err)
		return
	}
	Broadcast(jsonMsg)
}

// func mainTest() {
// 	http.HandleFunc("/ws", HandleWebSocket)
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		http.ServeFile(w, r, "index.html")
// 	})
// 	go SendMsg()
// 	http.ListenAndServe(":8080", nil)
// }

func SendMsg() {
	for {
		time.Sleep(1 * time.Second)
		fmt.Println("xiao")
		msg := Msg{Data: "xiaojin", Action: "pop"}
		msgJson, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("ok")
			return
		}
		Broadcast(msgJson)
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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
		SendJsonMsg(string(msg), "process")
	}
}

func Broadcast(msg []byte) {
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println(err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
