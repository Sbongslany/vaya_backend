package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type AcceptTripOfferInput struct {
	TripID      uuid.UUID
	OfferID     uuid.UUID
	PassengerID uuid.UUID
}

type AcceptTripOffer struct {
	tripRepo      repositories.TripRepository
	tripOfferRepo repositories.TripOfferRepository
	stateMachine  *services.StateMachine
}

func NewAcceptTripOffer(
	tripRepo repositories.TripRepository,
	tripOfferRepo repositories.TripOfferRepository,
	stateMachine *services.StateMachine,
) *AcceptTripOffer {
	return &AcceptTripOffer{
		tripRepo:      tripRepo,
		tripOfferRepo: tripOfferRepo,
		stateMachine:  stateMachine,
	}
}

func (uc *AcceptTripOffer) Execute(ctx context.Context, input AcceptTripOfferInput) (*entities.Trip, error) {
	// 1. Get the trip
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	// 2. Verify passenger owns this trip
	if trip.PassengerID != input.PassengerID {
		return nil, domain.ErrUnauthorized
	}

	// 3. Validate state transition
	if err := uc.stateMachine.Transition(trip.Status, entities.StatusDriverAssigned); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	// 4. Get the offer
	offer, err := uc.tripOfferRepo.GetByID(ctx, input.OfferID)
	if err != nil {
		return nil, err
	}
	if offer == nil {
		return nil, domain.ErrOfferNotFound
	}

	// 5. Verify offer belongs to this trip and is still pending
	if offer.TripID != input.TripID {
		return nil, domain.ErrOfferNotFound
	}
	if offer.Status != entities.OfferStatusPending {
		return nil, domain.ErrInvalidStateTransition
	}

	// 6. Assign driver to trip (most critical operation first)
	if err := uc.tripRepo.AssignDriver(ctx, input.TripID, offer.DriverID, entities.StatusDriverAssigned); err != nil {
		return nil, err
	}

	// 7. Mark offer as accepted
	if err := uc.tripOfferRepo.UpdateStatus(ctx, offer.ID, entities.OfferStatusAccepted); err != nil {
		return nil, err
	}

	// 8. Reject all other pending offers
	if err := uc.tripOfferRepo.RejectOthersForTrip(ctx, input.TripID, offer.ID); err != nil {
		return nil, err
	}

	trip.DriverID = &offer.DriverID
	trip.Status = entities.StatusDriverAssigned
	return trip, nil
}
