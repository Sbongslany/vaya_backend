package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type WaypointRepository interface {
	CreateMany(ctx context.Context, waypoints []*entities.Waypoint) error
	FindByTripID(ctx context.Context, tripID uuid.UUID) ([]*entities.Waypoint, error)
}
