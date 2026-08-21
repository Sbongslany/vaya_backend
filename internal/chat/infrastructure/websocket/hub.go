package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	UserID uuid.UUID
	TripID uuid.UUID
	Conn   *websocket.Conn
}

type ChatHub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID][]*Client // Keyed by TripID
}

func NewChatHub() *ChatHub {
	return &ChatHub{
		clients: make(map[uuid.UUID][]*Client),
	}
}

func (h *ChatHub) ServeWS(w http.ResponseWriter, r *http.Request, userID, tripID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		UserID: userID,
		TripID: tripID,
		Conn:   conn,
	}

	h.mu.Lock()
	h.clients[tripID] = append(h.clients[tripID], client)
	h.mu.Unlock()

	// Keep connection alive and clean up on disconnect
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				h.mu.Lock()
				clients := h.clients[tripID]
				for i, c := range clients {
					if c == client {
						h.clients[tripID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				h.mu.Unlock()
				conn.Close()
				return
			}
		}
	}()
}

// Broadcast sends a payload to all connected users in a specific trip
func (h *ChatHub) Broadcast(tripID uuid.UUID, eventType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := map[string]interface{}{
		"event": eventType,
		"data":  payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for _, client := range h.clients[tripID] {
		_ = client.Conn.WriteMessage(websocket.TextMessage, data)
	}
}
