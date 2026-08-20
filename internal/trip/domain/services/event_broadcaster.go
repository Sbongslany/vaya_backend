package services

import (
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

// EventBroadcaster pushes trip events to connected real-time clients.
type EventBroadcaster interface {
	Broadcast(tripID uuid.UUID, event *entities.TripEvent)
}