package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ScheduleLongDistanceTripInput struct {
	TripID      uuid.UUID
	PassengerID uuid.UUID
}

type ScheduleLongDistanceTrip struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewScheduleLongDistanceTrip(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *ScheduleLongDistanceTrip {
	return &ScheduleLongDistanceTrip{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *ScheduleLongDistanceTrip) Execute(ctx context.Context, input ScheduleLongDistanceTripInput) (*entities.Trip, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	if trip.PassengerID != input.PassengerID {
		return nil, domain.ErrUnauthorized
	}

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusScheduled); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusScheduled); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusScheduled
	return trip, nil
}
