package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type DepartForPickupInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type DepartForPickup struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewDepartForPickup(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *DepartForPickup {
	return &DepartForPickup{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *DepartForPickup) Execute(ctx context.Context, input DepartForPickupInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusDriverEnRoute); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusDriverEnRoute); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusDriverEnRoute
	return trip, nil
}
