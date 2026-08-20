package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type CancelTripInput struct {
	TripID uuid.UUID
	UserID uuid.UUID
	Reason string
}

type CancelTrip struct {
	tripRepo      repositories.TripRepository
	tripOfferRepo repositories.TripOfferRepository
	stateMachine  *services.StateMachine
}

func NewCancelTrip(
	tripRepo repositories.TripRepository,
	tripOfferRepo repositories.TripOfferRepository,
	stateMachine *services.StateMachine,
) *CancelTrip {
	return &CancelTrip{
		tripRepo:      tripRepo,
		tripOfferRepo: tripOfferRepo,
		stateMachine:  stateMachine,
	}
}

func (uc *CancelTrip) Execute(ctx context.Context, input CancelTripInput) (*entities.Trip, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	// Determine who is cancelling and the target status
	var cancelStatus entities.TripStatus
	switch {
	case input.UserID == trip.PassengerID:
		cancelStatus = entities.StatusCancelledByPassenger
	case trip.DriverID != nil && input.UserID == *trip.DriverID:
		cancelStatus = entities.StatusCancelledByDriver
	default:
		return nil, domain.ErrUnauthorized
	}

	// Validate the transition is allowed from the current state
	if err := uc.stateMachine.Transition(trip.Status, cancelStatus); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	// Calculate cancellation fee
	fee := uc.calculateCancellationFee(trip)

	now := time.Now()
	trip.Status = cancelStatus
	trip.CancelledBy = &input.UserID
	trip.CancelledAt = &now
	trip.CancellationFee = fee
	if input.Reason != "" {
		trip.CancellationReason = &input.Reason
	}

	if err := uc.tripRepo.Cancel(ctx, trip); err != nil {
		return nil, err
	}

	// Expire any pending offers for this trip
	if err := uc.tripOfferRepo.ExpireAllForTrip(ctx, trip.ID); err != nil {
		return nil, err
	}

	return trip, nil
}

// calculateCancellationFee applies the cancellation policy:
// - No fee if no driver was assigned yet.
// - Flat fee if a driver had already been assigned.
func (uc *CancelTrip) calculateCancellationFee(trip *entities.Trip) *float64 {
	if trip.DriverID == nil {
		return nil
	}
	fee := 20.0
	return &fee
}
