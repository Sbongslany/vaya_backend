package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ArriveAtPickupInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type ArriveAtPickup struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewArriveAtPickup(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *ArriveAtPickup {
	return &ArriveAtPickup{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *ArriveAtPickup) Execute(ctx context.Context, input ArriveAtPickupInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusDriverArrived); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusDriverArrived); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusDriverArrived
	return trip, nil
}
