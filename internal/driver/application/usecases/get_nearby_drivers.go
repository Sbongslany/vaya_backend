package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/driver/domain/repositories"
)

type GetNearbyDrivers struct {
	locationRepo repositories.DriverLocationRepository
}

func NewGetNearbyDrivers(locationRepo repositories.DriverLocationRepository) *GetNearbyDrivers {
	return &GetNearbyDrivers{locationRepo: locationRepo}
}

func (uc *GetNearbyDrivers) Execute(ctx context.Context, lat, lng, radiusKM float64) ([]string, error) {
	if radiusKM <= 0 {
		radiusKM = 5.0 // Default 5km
	}
	return uc.locationRepo.FindNearbyDrivers(ctx, lat, lng, radiusKM)
}
