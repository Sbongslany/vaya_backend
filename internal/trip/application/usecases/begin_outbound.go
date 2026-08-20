package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type BeginOutboundInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type BeginOutbound struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewBeginOutbound(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *BeginOutbound {
	return &BeginOutbound{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *BeginOutbound) Execute(ctx context.Context, input BeginOutboundInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusOutboundInProgress); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusOutboundInProgress); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusOutboundInProgress
	return trip, nil
}
