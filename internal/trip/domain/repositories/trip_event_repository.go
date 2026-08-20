package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type TripEventRepository interface {
	Create(ctx context.Context, event *entities.TripEvent) error
	FindByTripID(ctx context.Context, tripID uuid.UUID) ([]*entities.TripEvent, error)
}
