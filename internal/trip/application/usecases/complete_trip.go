package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type CompleteTripInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type CompleteTrip struct {
	tripRepo           repositories.TripRepository
	stateMachine       *services.StateMachine
	driverStateManager services.DriverStateManager
	fareSplitter       services.FareSplitter
}

func NewCompleteTrip(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
	driverStateManager services.DriverStateManager,
	fareSplitter services.FareSplitter,
) *CompleteTrip {
	return &CompleteTrip{
		tripRepo:           tripRepo,
		stateMachine:       stateMachine,
		driverStateManager: driverStateManager,
		fareSplitter:       fareSplitter,
	}
}

func (uc *CompleteTrip) Execute(ctx context.Context, input CompleteTripInput) (*entities.Trip, error) {
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

	// Use final fare if set, otherwise use estimated fare
	finalFare := trip.EstimatedFare
	if trip.FinalFare != nil {
		finalFare = *trip.FinalFare
	}

	if err := uc.tripRepo.UpdateStatusAndFinalFare(ctx, trip.ID, entities.StatusTripCompleted, finalFare); err != nil {
		return nil, err
	}

	// Mark driver as ONLINE so they can receive new trips
	if trip.DriverID != nil && uc.driverStateManager != nil {
		_ = uc.driverStateManager.MarkOnline(ctx, trip.DriverID.String())
	}

	// Automatically split the fare into wallets (Driver gets 80%, Platform gets 20%)
	if uc.fareSplitter != nil && finalFare > 0 {
		_ = uc.fareSplitter.SplitFare(ctx, trip.ID, trip.PassengerID, input.DriverID, finalFare)
	}

	trip.Status = entities.StatusTripCompleted
	trip.FinalFare = &finalFare
	return trip, nil
}
