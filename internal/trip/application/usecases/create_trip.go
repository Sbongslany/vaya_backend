package usecases

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type CreateTripInput struct {
	PassengerID      uuid.UUID
	PickupLatitude   float64
	PickupLongitude  float64
	PickupAddress    string
	DropoffLatitude  float64
	DropoffLongitude float64
	DropoffAddress   string
}

type CreateTrip struct {
	tripRepo     repositories.TripRepository
	fareCalc     *services.FareCalculator
	eventService *services.TripEventService
}

func NewCreateTrip(
	tripRepo repositories.TripRepository,
	fareCalc *services.FareCalculator,
	eventService *services.TripEventService,
) *CreateTrip {
	return &CreateTrip{
		tripRepo:     tripRepo,
		fareCalc:     fareCalc,
		eventService: eventService,
	}
}

func (uc *CreateTrip) Execute(ctx context.Context, input CreateTripInput) (*entities.Trip, error) {
	if err := validateCoordinates(input.PickupLatitude, input.PickupLongitude); err != nil {
		return nil, err
	}
	if err := validateCoordinates(input.DropoffLatitude, input.DropoffLongitude); err != nil {
		return nil, err
	}

	activeTrip, err := uc.tripRepo.FindActiveByPassengerID(ctx, input.PassengerID)
	if err != nil {
		return nil, err
	}
	if activeTrip != nil {
		return nil, domain.ErrActiveTripExists
	}

	distanceKM := services.CalculateDistanceKM(
		input.PickupLatitude, input.PickupLongitude,
		input.DropoffLatitude, input.DropoffLongitude,
	)
	estimatedMinutes := int(distanceKM * 2)
	estimatedFare := uc.fareCalc.Calculate(distanceKM, estimatedMinutes)

	now := time.Now()
	trip := &entities.Trip{
		ID:               uuid.New(),
		PassengerID:      input.PassengerID,
		TripType:         entities.TripTypeNormal,
		Status:           entities.StatusRequested,
		StartPIN:         generatePIN(),
		PickupLatitude:   input.PickupLatitude,
		PickupLongitude:  input.PickupLongitude,
		PickupAddress:    input.PickupAddress,
		DropoffLatitude:  input.DropoffLatitude,
		DropoffLongitude: input.DropoffLongitude,
		DropoffAddress:   input.DropoffAddress,
		EstimatedFare:    estimatedFare,
		Currency:         "ZAR",
		DistanceKM:       &distanceKM,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := uc.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	// Record trip created event
	fromStatus := ""
	toStatus := string(entities.StatusRequested)
	_ = uc.eventService.Record(
		ctx,
		trip.ID,
		entities.EventTypeTripCreated,
		&input.PassengerID,
		fromStatus,
		toStatus,
		map[string]interface{}{
			"pickup_address":  input.PickupAddress,
			"dropoff_address": input.DropoffAddress,
			"estimated_fare":  estimatedFare,
			"distance_km":     distanceKM,
		},
	)

	return trip, nil
}

func validateCoordinates(lat, lng float64) error {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return domain.ErrInvalidCoordinates
	}
	return nil
}

func generatePIN() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "0000"
	}
	return fmt.Sprintf("%04d", n.Int64())
}