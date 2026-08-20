package entities

import "github.com/google/uuid"

type RatingSummary struct {
	UserID      uuid.UUID
	RatingAvg   float64
	RatingCount int
}
