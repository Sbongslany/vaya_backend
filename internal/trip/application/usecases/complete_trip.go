package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type CompleteTripInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type CompleteTrip struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewCompleteTrip(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *CompleteTrip {
	return &CompleteTrip{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *CompleteTrip) Execute(ctx context.Context, input CompleteTripInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusTripCompleted); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	// Use estimated fare as final fare (will be enhanced with GPS tracking later)
	finalFare := trip.EstimatedFare

	if err := uc.tripRepo.UpdateStatusAndFinalFare(ctx, trip.ID, entities.StatusTripCompleted, finalFare); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusTripCompleted
	trip.FinalFare = &finalFare
	return trip, nil
}
