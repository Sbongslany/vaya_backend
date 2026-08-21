package repositories

import (
	"context"
	"time"

	"github.com/yourorg/ehailing/backend/internal/driver/domain/entities"
)

type DriverStateRepository interface {
	SetStatus(ctx context.Context, driverID string, status entities.DriverStatus) error
	GetStatus(ctx context.Context, driverID string) (entities.DriverStatus, error)
	Delete(ctx context.Context, driverID string) error
}

type DriverLocationRepository interface {
	UpdateLocation(ctx context.Context, loc *entities.DriverLocation) error
	GetLocation(ctx context.Context, driverID string) (*entities.DriverLocation, error)
	// FindNearbyDrivers returns driver IDs within the radius (km) of the given point
	FindNearbyDrivers(ctx context.Context, lat, lng, radiusKM float64) ([]string, error)
	RemoveLocation(ctx context.Context, driverID string) error
	SetTTL(ctx context.Context, driverID string, ttl time.Duration) error
}
