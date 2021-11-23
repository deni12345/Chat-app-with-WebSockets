package main

import (
	"chatapp/model"
	"log"
	_ "log"
	"net/http"
	_ "net/http"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan model.Message)
	upgrader  = websocket.Upgrader{}
)

func main() {
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	go handleMessage()

	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer ws.Close()

	clients[ws] = true
	for {
		var msg model.Message

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

func handleMessage() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
