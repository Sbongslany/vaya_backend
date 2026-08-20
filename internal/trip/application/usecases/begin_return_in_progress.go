package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type BeginReturnInProgressInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type BeginReturnInProgress struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewBeginReturnInProgress(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *BeginReturnInProgress {
	return &BeginReturnInProgress{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *BeginReturnInProgress) Execute(ctx context.Context, input BeginReturnInProgressInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusReturnInProgress); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusReturnInProgress); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusReturnInProgress
	return trip, nil
}
