package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type GetTripHistoryInput struct {
	TripID uuid.UUID
	UserID uuid.UUID
}

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

func (uc *GetTripHistory) Execute(ctx context.Context, input GetTripHistoryInput) ([]*entities.TripEvent, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	// Only trip participants can view the event history
	isParticipant := input.UserID == trip.PassengerID ||
		(trip.DriverID != nil && input.UserID == *trip.DriverID)
	if !isParticipant {
		return nil, domain.ErrUnauthorized
	}

	return uc.eventRepo.FindByTripID(ctx, input.TripID)
}
