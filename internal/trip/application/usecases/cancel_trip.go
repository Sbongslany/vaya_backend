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
	eventService  *services.TripEventService
}

func NewCancelTrip(
	tripRepo repositories.TripRepository,
	tripOfferRepo repositories.TripOfferRepository,
	stateMachine *services.StateMachine,
	eventService *services.TripEventService,
) *CancelTrip {
	return &CancelTrip{
		tripRepo:      tripRepo,
		tripOfferRepo: tripOfferRepo,
		stateMachine:  stateMachine,
		eventService:  eventService,
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

	var cancelStatus entities.TripStatus
	switch {
	case input.UserID == trip.PassengerID:
		cancelStatus = entities.StatusCancelledByPassenger
	case trip.DriverID != nil && input.UserID == *trip.DriverID:
		cancelStatus = entities.StatusCancelledByDriver
	default:
		return nil, domain.ErrUnauthorized
	}

	if err := uc.stateMachine.Transition(trip.Status, cancelStatus); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	fee := uc.calculateCancellationFee(trip)

	// Capture old status before updating
	oldStatus := string(trip.Status)

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

	if err := uc.tripOfferRepo.ExpireAllForTrip(ctx, trip.ID); err != nil {
		return nil, err
	}

	// Record cancellation event
	payload := map[string]interface{}{
		"status": string(cancelStatus),
	}
	if input.Reason != "" {
		payload["reason"] = input.Reason
	}
	if fee != nil {
		payload["cancellation_fee"] = *fee
	}

	newStatus := string(cancelStatus)
	_ = uc.eventService.Record(ctx, trip.ID, entities.EventTypeTripCancelled, &input.UserID, oldStatus, newStatus, payload)

	return trip, nil
}

func (uc *CancelTrip) calculateCancellationFee(trip *entities.Trip) *float64 {
	if trip.DriverID == nil {
		return nil
	}
	fee := 20.0
	return &fee
}
