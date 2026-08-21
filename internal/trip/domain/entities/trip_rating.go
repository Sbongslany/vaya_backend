package entities

import (
	"time"

	"github.com/google/uuid"
)

type TripRating struct {
	ID          uuid.UUID
	TripID      uuid.UUID
	RaterID     uuid.UUID
	RatedUserID uuid.UUID
	Rating      int
	Comment     string
	CreatedAt   time.Time
}
