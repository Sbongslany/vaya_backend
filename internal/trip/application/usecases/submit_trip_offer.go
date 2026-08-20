package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type SubmitTripOfferInput struct {
	TripID      uuid.UUID
	DriverID    uuid.UUID
	OfferType   entities.OfferType
	OfferedFare float64
}

type SubmitTripOffer struct {
	tripRepo      repositories.TripRepository
	tripOfferRepo repositories.TripOfferRepository
}

func NewSubmitTripOffer(tripRepo repositories.TripRepository, tripOfferRepo repositories.TripOfferRepository) *SubmitTripOffer {
	return &SubmitTripOffer{
		tripRepo:      tripRepo,
		tripOfferRepo: tripOfferRepo,
	}
}

func (uc *SubmitTripOffer) Execute(ctx context.Context, input SubmitTripOfferInput) (*entities.TripOffer, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	// Only allow offers on trips that are REQUESTED or SEARCHING_DRIVERS
	if trip.Status != entities.StatusRequested && trip.Status != entities.StatusSearchingDrivers {
		return nil, domain.ErrInvalidStateTransition
	}

	if input.OfferedFare <= 0 {
		return nil, domain.ErrInvalidOfferFare
	}

	now := time.Now()
	offer := &entities.TripOffer{
		ID:          uuid.New(),
		TripID:      input.TripID,
		DriverID:    input.DriverID,
		OfferType:   input.OfferType,
		OfferedFare: input.OfferedFare,
		Status:      entities.OfferStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.tripOfferRepo.Create(ctx, offer); err != nil {
		return nil, err
	}

	return offer, nil
}