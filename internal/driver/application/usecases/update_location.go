package usecases

import (
	"context"
	"math"
	"time"

	"github.com/yourorg/ehailing/backend/internal/driver/domain"
	"github.com/yourorg/ehailing/backend/internal/driver/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/driver/domain/repositories"
)

type UpdateLocationInput struct {
	DriverID  string
	Latitude  float64
	Longitude float64
	Heading   float64
	Speed     float64
}

type UpdateLocation struct {
	locationRepo repositories.DriverLocationRepository
	stateRepo    repositories.DriverStateRepository
}

func NewUpdateLocation(
	locationRepo repositories.DriverLocationRepository,
	stateRepo repositories.DriverStateRepository,
) *UpdateLocation {
	return &UpdateLocation{
		locationRepo: locationRepo,
		stateRepo:    stateRepo,
	}
}

func (uc *UpdateLocation) Execute(ctx context.Context, input UpdateLocationInput) error {
	if input.Latitude < -90 || input.Latitude > 90 || input.Longitude < -180 || input.Longitude > 180 {
		return domain.ErrInvalidLocation
	}

	// Only update location if driver is ONLINE or BUSY
	status, err := uc.stateRepo.GetStatus(ctx, input.DriverID)
	if err != nil {
		return err
	}
	if status == entities.DriverStatusOffline {
		return nil // Silently ignore pings from offline drivers
	}

	loc := &entities.DriverLocation{
		DriverID:  input.DriverID,
		Latitude:  math.Round(input.Latitude*1e6) / 1e6, // 6 decimal places
		Longitude: math.Round(input.Longitude*1e6) / 1e6,
		Heading:   input.Heading,
		Speed:     input.Speed,
		UpdatedAt: time.Now(),
	}

	return uc.locationRepo.UpdateLocation(ctx, loc)
}
