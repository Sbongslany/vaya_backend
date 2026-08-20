package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type StartTripInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
	PIN      string
}

type StartTrip struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewStartTrip(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *StartTrip {
	return &StartTrip{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *StartTrip) Execute(ctx context.Context, input StartTripInput) (*entities.Trip, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	if trip.DriverID == nil || *trip.DriverID != input.DriverID {
		return nil, domain.ErrUnauthorized
	}

	// Verify PIN before starting the trip
	if trip.StartPIN != input.PIN {
		return nil, domain.ErrInvalidPIN
	}

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusTripInProgress); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusTripInProgress); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusTripInProgress
	return trip, nil
}
