package entities

import (
	"time"

	"github.com/google/uuid"
)

type PromotionRedemption struct {
	ID              uuid.UUID
	PromotionID     uuid.UUID
	UserID          uuid.UUID
	TripID          *uuid.UUID
	DiscountApplied float64
	RedeemedAt      time.Time
}
