package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeTripCreated    = "TRIP_CREATED"
	EventTypeStatusChanged  = "STATUS_CHANGED"
	EventTypeDriverAssigned = "DRIVER_ASSIGNED"
	EventTypeTripCancelled  = "TRIP_CANCELLED"
)

type TripEvent struct {
	ID         uuid.UUID
	TripID     uuid.UUID
	EventType  string
	ActorID    *uuid.UUID
	FromStatus *string
	ToStatus   *string
	Metadata   json.RawMessage
	CreatedAt  time.Time
}
