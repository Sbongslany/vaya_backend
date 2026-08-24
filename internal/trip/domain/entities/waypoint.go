package entities

import (
	"time"

	"github.com/google/uuid"
)

type Waypoint struct {
	ID        uuid.UUID
	TripID    uuid.UUID
	Sequence  int
	Latitude  float64
	Longitude float64
	Address   string
	CreatedAt time.Time
}
