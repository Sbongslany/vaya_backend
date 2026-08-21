package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type TripRatingRepository interface {
	Create(ctx context.Context, rating *entities.TripRating) error
	FindByTripAndRater(ctx context.Context, tripID uuid.UUID, raterID uuid.UUID) (*entities.TripRating, error)
	CountByTripID(ctx context.Context, tripID uuid.UUID) (int, error)
}
