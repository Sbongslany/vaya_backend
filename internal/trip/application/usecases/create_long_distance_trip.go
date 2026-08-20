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

type CreateLongDistanceTripInput struct {
	PassengerID        uuid.UUID
	PickupLatitude     float64
	PickupLongitude    float64
	PickupAddress      string
	DropoffLatitude    float64
	DropoffLongitude   float64
	DropoffAddress     string
	LongDistanceType   entities.LongDistanceType
	ScheduledDeparture time.Time
	ScheduledReturn    *time.Time
	TripDurationDays   int
}

type CreateLongDistanceTrip struct {
	tripRepo repositories.TripRepository
	fareCalc *services.FareCalculator
}

func NewCreateLongDistanceTrip(tripRepo repositories.TripRepository, fareCalc *services.FareCalculator) *CreateLongDistanceTrip {
	return &CreateLongDistanceTrip{
		tripRepo: tripRepo,
		fareCalc: fareCalc,
	}
}

func (uc *CreateLongDistanceTrip) Execute(ctx context.Context, input CreateLongDistanceTripInput) (*entities.Trip, error) {
	if err := validateCoordinates(input.PickupLatitude, input.PickupLongitude); err != nil {
		return nil, err
	}
	if err := validateCoordinates(input.DropoffLatitude, input.DropoffLongitude); err != nil {
		return nil, err
	}

	// Validate schedule
	if input.ScheduledDeparture.Before(time.Now()) {
		return nil, domain.ErrInvalidSchedule
	}
	switch input.LongDistanceType {
	case entities.LongDistanceOneWay:
		// No extra requirements
	case entities.LongDistanceReturn, entities.LongDistanceMultiDay:
		if input.ScheduledReturn == nil || !input.ScheduledReturn.After(input.ScheduledDeparture) {
			return nil, domain.ErrInvalidSchedule
		}
	default:
		return nil, domain.ErrInvalidSchedule
	}
	if input.LongDistanceType == entities.LongDistanceMultiDay && input.TripDurationDays <= 0 {
		return nil, domain.ErrInvalidSchedule
	}

	// Prevent duplicate active trips
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
	estimatedFare := uc.fareCalc.CalculateLongDistanceFare(distanceKM, estimatedMinutes, input.LongDistanceType, input.TripDurationDays)

	now := time.Now()
	trip := &entities.Trip{
		ID:                 uuid.New(),
		PassengerID:        input.PassengerID,
		TripType:           entities.TripTypeLongDistance,
		Status:             entities.StatusQuoteGenerated,
		StartPIN:           generatePIN(),
		PickupLatitude:     input.PickupLatitude,
		PickupLongitude:    input.PickupLongitude,
		PickupAddress:      input.PickupAddress,
		DropoffLatitude:    input.DropoffLatitude,
		DropoffLongitude:   input.DropoffLongitude,
		DropoffAddress:     input.DropoffAddress,
		EstimatedFare:      estimatedFare,
		Currency:           "ZAR",
		DistanceKM:         &distanceKM,
		LongDistanceType:   &input.LongDistanceType,
		ScheduledDeparture: &input.ScheduledDeparture,
		ScheduledReturn:    input.ScheduledReturn,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if input.LongDistanceType == entities.LongDistanceMultiDay {
		trip.TripDurationDays = &input.TripDurationDays
	}

	if err := uc.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}
