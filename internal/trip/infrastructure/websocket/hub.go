package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins for dev
}

type Client struct {
	UserID uuid.UUID
	TripID uuid.UUID
	Conn   *websocket.Conn
}

type Hub struct {
	mu      sync.RWMutex
	clients []*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make([]*Client, 0),
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, userID, tripID uuid.UUID) {
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
	h.clients = append(h.clients, client)
	h.mu.Unlock()

	// Keep connection alive and clean up on disconnect
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				h.mu.Lock()
				for i, c := range h.clients {
					if c == client {
						h.clients = append(h.clients[:i], h.clients[i+1:]...)
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

func (h *Hub) Broadcast(tripID uuid.UUID, event *entities.TripEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	payload, _ := json.Marshal(event)
	msg := map[string]interface{}{
		"trip_id": tripID,
		"event":   json.RawMessage(payload),
	}
	data, _ := json.Marshal(msg)

	for _, client := range h.clients {
		if client.TripID == tripID {
			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}
