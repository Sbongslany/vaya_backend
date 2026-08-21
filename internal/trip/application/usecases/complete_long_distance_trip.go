package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type CompleteLongDistanceTripInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type CompleteLongDistanceTrip struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
	fareSplitter services.FareSplitter
}

func NewCompleteLongDistanceTrip(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
	fareSplitter services.FareSplitter,
) *CompleteLongDistanceTrip {
	return &CompleteLongDistanceTrip{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
		fareSplitter: fareSplitter,
	}
}

func (uc *CompleteLongDistanceTrip) Execute(ctx context.Context, input CompleteLongDistanceTripInput) (*entities.Trip, error) {
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

	if err := uc.stateMachine.Transition(trip.Status, entities.StatusTripCompleted); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	finalFare := trip.EstimatedFare
	if trip.FinalFare != nil {
		finalFare = *trip.FinalFare
	}

	if err := uc.tripRepo.UpdateStatusAndFinalFare(ctx, trip.ID, entities.StatusTripCompleted, finalFare); err != nil {
		return nil, err
	}

	// Automatically split the fare into wallets
	if uc.fareSplitter != nil && finalFare > 0 {
		_ = uc.fareSplitter.SplitFare(ctx, trip.ID, trip.PassengerID, input.DriverID, finalFare)
	}

	trip.Status = entities.StatusTripCompleted
	trip.FinalFare = &finalFare
	return trip, nil
}
