package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/driver/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/driver/domain/repositories"
)

type GoOffline struct {
	stateRepo    repositories.DriverStateRepository
	locationRepo repositories.DriverLocationRepository
}

func NewGoOffline(
	stateRepo repositories.DriverStateRepository,
	locationRepo repositories.DriverLocationRepository,
) *GoOffline {
	return &GoOffline{
		stateRepo:    stateRepo,
		locationRepo: locationRepo,
	}
}

func (uc *GoOffline) Execute(ctx context.Context, driverID string) error {
	// Set status to offline
	if err := uc.stateRepo.SetStatus(ctx, driverID, entities.DriverStatusOffline); err != nil {
		return err
	}

	// Remove from the live map so they stop appearing in "nearby drivers"
	_ = uc.locationRepo.RemoveLocation(ctx, driverID)

	return nil
}
