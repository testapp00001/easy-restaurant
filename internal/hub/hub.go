package hub

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Client represents a WebSocket client
type Client struct {
	Conn *websocket.Conn
	Room string
}

// Message represents a message to be broadcasted
type Message struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	rooms      map[string]map[*websocket.Conn]bool
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*websocket.Conn]bool),
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Register adds a new client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.Room] == nil {
				h.rooms[client.Room] = make(map[*websocket.Conn]bool)
			}
			h.rooms[client.Room][client.Conn] = true
			h.clients[client.Conn] = true
			log.Printf("Client registered to room: %s", client.Room)
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.rooms[client.Room]; ok {
				delete(h.rooms[client.Room], client.Conn)
				if len(h.rooms[client.Room]) == 0 {
					delete(h.rooms, client.Room)
				}
			}
			delete(h.clients, client.Conn)
			log.Printf("Client unregistered from room: %s", client.Room)
			h.mu.Unlock()
		}
	}
}

// BroadcastToRoom sends a message to all clients in a specific room.
func (h *Hub) BroadcastToRoom(room string, event string, data interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	messageBytes, err := json.Marshal(Message{Event: event, Data: data})
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for conn := range h.rooms[room] {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("Error writing message to client: %v", err)
		}
	}
}
