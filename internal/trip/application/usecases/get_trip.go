package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type GetTrip struct {
	tripRepo repositories.TripRepository
}

func NewGetTrip(tripRepo repositories.TripRepository) *GetTrip {
	return &GetTrip{tripRepo: tripRepo}
}

func (uc *GetTrip) Execute(ctx context.Context, tripID uuid.UUID) (*entities.Trip, error) {
	trip, err := uc.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}
	return trip, nil
}