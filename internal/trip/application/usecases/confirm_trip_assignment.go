package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ConfirmTripAssignmentInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type ConfirmTripAssignment struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewConfirmTripAssignment(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *ConfirmTripAssignment {
	return &ConfirmTripAssignment{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *ConfirmTripAssignment) Execute(ctx context.Context, input ConfirmTripAssignmentInput) (*entities.Trip, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	// Verify this driver is the assigned driver
	if trip.DriverID == nil || *trip.DriverID != input.DriverID {
		return nil, domain.ErrUnauthorized
	}

	// Validate state transition
	if err := uc.stateMachine.Transition(trip.Status, entities.StatusDriverEnRoute); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	// Update status
	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusDriverEnRoute); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusDriverEnRoute
	return trip, nil
}
