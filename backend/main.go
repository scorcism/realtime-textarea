package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]string)
	broadcast = make(chan Message)
	mutex     sync.Mutex
)

type Message struct {
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CursorPos CursorPos `json:"cursorPos"`
}

var user struct {
	Username string `json:"username"`
}

type CursorPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var messages = ""

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	if err := ws.ReadJSON(&user); err != nil {
		log.Println("Error getting username:", err)
		return
	}

	mutex.Lock()
	clients[ws] = user.Username
	mutex.Unlock()

	sendLatestDocument(ws)

	for {
		var msg Message
		if err := ws.ReadJSON(&msg); err != nil {
			log.Println("Error reading message:", err)
			mutex.Lock()
			delete(clients, ws)
			mutex.Unlock()
			break
		}

		if msg.Type == "text" {
			updateDocument(msg.Content)
		}

		broadcast <- msg
	}
}

func sendLatestDocument(ws *websocket.Conn) {
	log.Println("Messages: ", messages)
	ws.WriteJSON(Message{
		Type:    "text",
		Content: messages,
	})
}

func updateDocument(content string) {
	log.Println("content: ", content)
	messages += content
}

func handleMessages() {
	for msg := range broadcast {
		mutex.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println("Error broadcasting message:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func getDocument(w http.ResponseWriter, r *http.Request) {

	json.NewEncoder(w).Encode(map[string]string{"content": messages})
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleConnections(w, r)
	})

	go handleMessages()

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
