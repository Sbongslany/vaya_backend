package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type PublishLongDistanceTripInput struct {
	TripID      uuid.UUID
	PassengerID uuid.UUID
}

type PublishLongDistanceTrip struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewPublishLongDistanceTrip(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *PublishLongDistanceTrip {
	return &PublishLongDistanceTrip{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *PublishLongDistanceTrip) Execute(ctx context.Context, input PublishLongDistanceTripInput) (*entities.Trip, error) {
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

	if trip.TripType != entities.TripTypeLongDistance {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusSearchingDrivers); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusSearchingDrivers); err != nil {
		return nil, err
	}

	trip.Status = entities.StatusSearchingDrivers
	return trip, nil
}
