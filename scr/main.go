package main

import (
	"chatapp/model"
	"log"
	_ "log"
	"net/http"
	_ "net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan model.Message)
	upgrader  = websocket.Upgrader{}
)

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../public")))
	d := http.Dir("../public")
	log.Println(d)

	r.HandleFunc("/ws", handleConnections)

	go handleMessage()

	log.Fatal(http.ListenAndServe(":8080", r))
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
