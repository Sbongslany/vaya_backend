package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type TripRepository interface {
	Create(ctx context.Context, trip *entities.Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Trip, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.TripStatus) error
	UpdateStatusAndFinalFare(ctx context.Context, id uuid.UUID, status entities.TripStatus, finalFare float64) error
	FindNearbyRequested(ctx context.Context, lat, lng, radiusKM float64, limit int) ([]*entities.Trip, error)
	FindActiveByPassengerID(ctx context.Context, passengerID uuid.UUID) (*entities.Trip, error)
	AssignDriver(ctx context.Context, tripID, driverID uuid.UUID, status entities.TripStatus) error
	FindOpenLongDistanceTrips(ctx context.Context, limit int) ([]*entities.Trip, error)
}

type TripOfferRepository interface {
	Create(ctx context.Context, offer *entities.TripOffer) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.TripOffer, error)
	FindByTripID(ctx context.Context, tripID uuid.UUID) ([]*entities.TripOffer, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.OfferStatus) error
	RejectOthersForTrip(ctx context.Context, tripID, exceptOfferID uuid.UUID) error
}
