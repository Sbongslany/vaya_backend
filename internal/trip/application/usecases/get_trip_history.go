package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type GetTripHistory struct {
	tripRepo  repositories.TripRepository
	eventRepo repositories.TripEventRepository
}

func NewGetTripHistory(
	tripRepo repositories.TripRepository,
	eventRepo repositories.TripEventRepository,
) *GetTripHistory {
	return &GetTripHistory{
		tripRepo:  tripRepo,
		eventRepo: eventRepo,
	}
}

func (uc *GetTripHistory) Execute(ctx context.Context, tripID uuid.UUID, requesterID uuid.UUID) ([]*entities.TripEvent, error) {
	trip, err := uc.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	isPassenger := trip.PassengerID == requesterID
	isDriver := trip.DriverID != nil && *trip.DriverID == requesterID
	if !isPassenger && !isDriver {
		return nil, domain.ErrUnauthorized
	}

	return uc.eventRepo.FindByTripID(ctx, tripID)
}
