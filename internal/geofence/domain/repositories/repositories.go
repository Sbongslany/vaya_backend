package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/geofence/domain/entities"
)

type GeofenceRepository interface {
	Create(ctx context.Context, fence *entities.Geofence) error
	List(ctx context.Context, activeOnly bool) ([]*entities.Geofence, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Geofence, error)
	// FindZonesContainingPoint returns all active geofences that contain the given lat/lng
	FindZonesContainingPoint(ctx context.Context, lat, lng float64) ([]*entities.Geofence, error)
}

type ZoneAssignmentRepository interface {
	AssignDriver(ctx context.Context, assignment *entities.ZoneAssignment) error
	RemoveDriver(ctx context.Context, driverID, zoneID uuid.UUID) error
	GetDriverAssignments(ctx context.Context, driverID uuid.UUID) ([]*entities.ZoneAssignment, error)
}
