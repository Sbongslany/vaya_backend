package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ScheduleReturnInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type ScheduleReturn struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewScheduleReturn(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *ScheduleReturn {
	return &ScheduleReturn{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *ScheduleReturn) Execute(ctx context.Context, input ScheduleReturnInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusReturnScheduled); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusReturnScheduled); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusReturnScheduled
	return trip, nil
}