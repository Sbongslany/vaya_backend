package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type GetNearbyTripsInput struct {
	Latitude  float64
	Longitude float64
	RadiusKM  float64
	Limit     int
}

type GetNearbyTrips struct {
	tripRepo repositories.TripRepository
}

func NewGetNearbyTrips(tripRepo repositories.TripRepository) *GetNearbyTrips {
	return &GetNearbyTrips{tripRepo: tripRepo}
}

func (uc *GetNearbyTrips) Execute(ctx context.Context, input GetNearbyTripsInput) ([]*entities.Trip, error) {
	if input.RadiusKM <= 0 {
		input.RadiusKM = 10.0
	}
	if input.RadiusKM > 50.0 {
		input.RadiusKM = 50.0
	}
	if input.Limit <= 0 || input.Limit > 50 {
		input.Limit = 20
	}

	return uc.tripRepo.FindNearbyRequested(ctx, input.Latitude, input.Longitude, input.RadiusKM, input.Limit)
}