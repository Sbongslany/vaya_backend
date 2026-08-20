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

	// Long-distance trips go to DRIVER_SELECTED (driver must confirm); normal trips go to DRIVER_ASSIGNED
	targetStatus := entities.StatusDriverAssigned
	if trip.TripType == entities.TripTypeLongDistance {
		targetStatus = entities.StatusDriverSelected
	}

	if err := uc.stateMachine.Transition(trip.Status, targetStatus); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}

	offer, err := uc.tripOfferRepo.GetByID(ctx, input.OfferID)
	if err != nil {
		return nil, err
	}
	if offer == nil {
		return nil, domain.ErrOfferNotFound
	}
	if offer.TripID != input.TripID {
		return nil, domain.ErrOfferNotFound
	}
	if offer.Status != entities.OfferStatusPending {
		return nil, domain.ErrInvalidStateTransition
	}

	if err := uc.tripRepo.AssignDriver(ctx, input.TripID, offer.DriverID, targetStatus); err != nil {
		return nil, err
	}

	if err := uc.tripOfferRepo.UpdateStatus(ctx, offer.ID, entities.OfferStatusAccepted); err != nil {
		return nil, err
	}

	if err := uc.tripOfferRepo.RejectOthersForTrip(ctx, input.TripID, offer.ID); err != nil {
		return nil, err
	}

	trip.DriverID = &offer.DriverID
	trip.Status = targetStatus
	return trip, nil
}
