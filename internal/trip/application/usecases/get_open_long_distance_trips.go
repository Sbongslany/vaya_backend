package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type GetOpenLongDistanceTrips struct {
	tripRepo repositories.TripRepository
}

func NewGetOpenLongDistanceTrips(tripRepo repositories.TripRepository) *GetOpenLongDistanceTrips {
	return &GetOpenLongDistanceTrips{tripRepo: tripRepo}
}

func (uc *GetOpenLongDistanceTrips) Execute(ctx context.Context, limit int) ([]*entities.Trip, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return uc.tripRepo.FindOpenLongDistanceTrips(ctx, limit)
}
