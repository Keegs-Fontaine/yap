package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

const StaticAssetsDir = "static"

//go:embed static/*
var staticFiles embed.FS

type Message struct {
	From    *websocket.Conn
	Content string
}

type Room struct {
	Name     string
	Messages []Message
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func viewRoom(r *Room) []string {
	formattedMessages := []string{fmt.Sprintf("ROOM %s\n\n------------\n\n", r.Name)}

	for _, m := range r.Messages {
		formattedMessages = append(formattedMessages,
			fmt.Sprintf("========\nNAME: %s\nMESSAGE: %s\n========\n", m.From.RemoteAddr(), m.Content),
		)
	}

	return formattedMessages
}

func main() {
	rooms := []Room{
		{Name: "first", Messages: []Message{}},
		{Name: "second", Messages: []Message{}},
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Err! Couldn't connect to websocket!", err)
		}

		for {
			var message struct {
				RoomId  int
				Message string
			}

			if err := conn.ReadJSON(&message); err != nil {
				log.Println("Err! Couldn't parse JSON message!", err)
			}

			room := &rooms[message.RoomId]

			room.Messages = append(room.Messages, Message{
				From:    conn,
				Content: message.Message,
			})

			if err := conn.WriteJSON(map[string]int{"apple": 5, "lettuce": 7}); err != nil {
				log.Println("Err! Couldn't write JSON message!", err)
			}
		}
	})

	http.HandleFunc("/view/{id}", func(w http.ResponseWriter, r *http.Request) {
		roomPath := r.PathValue("id")
		roomIndex, err := strconv.Atoi(roomPath)

		fmt.Println(rooms)

		if err != nil {
			log.Println("Err! Couldn't get roomIndex", err)
			w.Write([]byte("Invalid Room ID"))
		}

		room := rooms[roomIndex]
		full := strings.Join(viewRoom(&room), "")

		w.Write(
			[]byte(full),
		)
	})

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.FS(staticFS))
	http.Handle("/", fs)

	fmt.Println("Server starting...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Err! Failed to start server!")
	}
}
