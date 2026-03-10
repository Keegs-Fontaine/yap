package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

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

type RoomListing struct {
	Name        string
	LastMessage string
	Date        string
	PFP         string
	Id          int
}

type MessageListing struct {
	Name    string
	Message string
	Date    string
	PFP     string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
		defer conn.Close()

		for {
			var message struct {
				RoomId  string
				Message string
			}

			if err := conn.ReadJSON(&message); err != nil {
				log.Println("Err! Couldn't parse JSON message!", err)
				break
			}

			roomIndex, err := strconv.Atoi(message.RoomId)
			if err != nil {
				log.Println("Err! Couldn't parse ascii integer!", err)
			}

			room := &rooms[roomIndex]

			room.Messages = append(room.Messages, Message{
				From:    conn,
				Content: message.Message,
			})

			wsWriter, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("Err! Can't get next IO writer!", err)
			}

			tmpl := template.Must(template.ParseFiles("./static/templates/message.html"))
			tmpl.Execute(wsWriter, []MessageListing{
				{Name: conn.RemoteAddr().String(), Message: message.Message, Date: "2-2-2", PFP: "static/images/SAMPLE-pfp-1.png"},
			})

			wsWriter.Close()
		}
	})

	http.HandleFunc("/view/{id}", func(w http.ResponseWriter, r *http.Request) {
		roomPath := r.PathValue("id")
		roomIndex, err := strconv.Atoi(roomPath)

		if err != nil {
			log.Println("Err! Couldn't get roomIndex", err)
			w.Write([]byte("Invalid Room ID"))
		}

		if roomIndex < len(rooms) {
			room := rooms[roomIndex]

			messages := []MessageListing{}

			for _, msg := range room.Messages {
				messages = append(messages, MessageListing{
					Name:    "wizard",
					PFP:     "static/images/SAMPLE-pfp-1.png",
					Date:    "10-22-24",
					Message: msg.Content,
				})
			}

			tmpl := template.Must(template.ParseFiles("./static/templates/message.html"))

			tmpl.Execute(w, messages)
		} else {
			http.Error(w, "Err! That Room Doesn't Exist!", http.StatusBadRequest)
		}
	})

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		roomListings := []RoomListing{
			{
				Name:        "wizard",
				LastMessage: "HEY DUDE WTF??? Where's my crystal ball????",
				Date:        "12/22/2003",
				PFP:         "/static/images/SAMPLE-pfp-1.png",
				Id:          0,
			},
			{
				Name:        "Some Name",
				LastMessage: "Hey I'm a generic message! Aren't I cooler than lorem ipsum???",
				Date:        "3/2/2006",
				PFP:         "/static/images/SAMPLE-pfp-2.png",
				Id:          1,
			},
			{
				Name:        "Dark Souls",
				LastMessage: "Rahhh I'm the dark souls guy",
				Date:        "5/23/2024",
				Id:          2,
			},
		}

		tmpl := template.Must(template.ParseFiles("./static/index.html"))
		tmpl.Execute(w, roomListings)
	})

	fmt.Println("Server starting...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Err! Failed to start server!")
	}
}
