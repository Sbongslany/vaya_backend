package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type UserRatingRepository interface {
	AddRating(ctx context.Context, userID uuid.UUID, rating int) error
	GetRatingSummary(ctx context.Context, userID uuid.UUID) (*entities.RatingSummary, error)
}
