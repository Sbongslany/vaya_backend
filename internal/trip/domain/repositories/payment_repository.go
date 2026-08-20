package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entities.Payment) error
	GetByTripID(ctx context.Context, tripID uuid.UUID) (*entities.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PaymentStatus) error
}

type TripRatingRepository interface {
	Create(ctx context.Context, rating *entities.TripRating) error
	FindByTripID(ctx context.Context, tripID uuid.UUID) ([]*entities.TripRating, error)
	FindByTripAndRater(ctx context.Context, tripID, raterID uuid.UUID) (*entities.TripRating, error)
	CountByTripID(ctx context.Context, tripID uuid.UUID) (int, error)
}
