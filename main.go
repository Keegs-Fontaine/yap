package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Room struct {
	Name string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Err! Couldn't connect to websocket!", err)
		}

		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Err! Couldn't read message", err)
			}

			fmt.Println(msgType, string(message))
		}
	})

	fmt.Println("Server starting...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Err! Failed to start server!")
	}
}
