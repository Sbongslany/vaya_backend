package entities

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID         uuid.UUID
	TripID     uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
	ReadAt     *time.Time
	CreatedAt  time.Time
}
