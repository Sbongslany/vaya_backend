package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ReachOutboundDestinationInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type ReachOutboundDestination struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewReachOutboundDestination(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *ReachOutboundDestination {
	return &ReachOutboundDestination{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *ReachOutboundDestination) Execute(ctx context.Context, input ReachOutboundDestinationInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusDestinationReached); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusDestinationReached); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusDestinationReached
	return trip, nil
}
