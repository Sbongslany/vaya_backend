package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ResolveOutboundArrivalInput struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type ResolveOutboundArrival struct {
	tripRepo     repositories.TripRepository
	stateMachine *services.StateMachine
}

func NewResolveOutboundArrival(
	tripRepo repositories.TripRepository,
	stateMachine *services.StateMachine,
) *ResolveOutboundArrival {
	return &ResolveOutboundArrival{
		tripRepo:     tripRepo,
		stateMachine: stateMachine,
	}
}

func (uc *ResolveOutboundArrival) Execute(ctx context.Context, input ResolveOutboundArrivalInput) (*entities.Trip, error) {
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

	if trip.Status != entities.StatusDestinationReached {
		return nil, domain.ErrInvalidStateTransition
	}

	ldType := entities.LongDistanceOneWay
	if trip.LongDistanceType != nil {
		ldType = *trip.LongDistanceType
	}

	if ldType == entities.LongDistanceOneWay {
		// ONE_WAY: complete the trip and set final fare
		if err := uc.stateMachine.Transition(trip.Status, entities.StatusTripCompleted); err != nil {
			return nil, domain.ErrInvalidStateTransition
		}
		finalFare := trip.EstimatedFare
		if err := uc.tripRepo.UpdateStatusAndFinalFare(ctx, trip.ID, entities.StatusTripCompleted, finalFare); err != nil {
			return nil, err
		}
		trip.Status = entities.StatusTripCompleted
		trip.FinalFare = &finalFare
		return trip, nil
	}

	// RETURN / MULTI_DAY: retain the driver for the return journey
	if err := uc.stateMachine.Transition(trip.Status, entities.StatusDriverRetained); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}
	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusDriverRetained); err != nil {
		return nil, err
	}
	trip.Status = entities.StatusDriverRetained
	return trip, nil
}
