package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type GetTripOffers struct {
	tripRepo      repositories.TripRepository
	tripOfferRepo repositories.TripOfferRepository
}

func NewGetTripOffers(tripRepo repositories.TripRepository, tripOfferRepo repositories.TripOfferRepository) *GetTripOffers {
	return &GetTripOffers{
		tripRepo:      tripRepo,
		tripOfferRepo: tripOfferRepo,
	}
}

func (uc *GetTripOffers) Execute(ctx context.Context, tripID uuid.UUID) ([]*entities.TripOffer, error) {
	trip, err := uc.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	offers, err := uc.tripOfferRepo.FindByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	return offers, nil
}
