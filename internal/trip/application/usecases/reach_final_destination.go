package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ReachFinalDestinationInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type ReachFinalDestination struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewReachFinalDestination(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *ReachFinalDestination {
	return &ReachFinalDestination{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *ReachFinalDestination) Execute(ctx context.Context, input ReachFinalDestinationInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusFinalDestination); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusFinalDestination); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusFinalDestination
	return trip, nil
}
