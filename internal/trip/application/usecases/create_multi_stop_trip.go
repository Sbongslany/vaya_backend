package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type WaypointInput struct {
	Sequence  int     `json:"sequence"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

type CreateMultiStopTripInput struct {
	PassengerID      uuid.UUID
	PickupLatitude   float64
	PickupLongitude  float64
	PickupAddress    string
	DropoffLatitude  float64
	DropoffLongitude float64
	DropoffAddress   string
	Waypoints        []WaypointInput
}

type CreateMultiStopTrip struct {
	tripRepo     repositories.TripRepository
	waypointRepo repositories.WaypointRepository
	fareCalc     *services.FareCalculator
}

func NewCreateMultiStopTrip(
	tripRepo repositories.TripRepository,
	waypointRepo repositories.WaypointRepository,
	fareCalc *services.FareCalculator,
) *CreateMultiStopTrip {
	return &CreateMultiStopTrip{
		tripRepo:     tripRepo,
		waypointRepo: waypointRepo,
		fareCalc:     fareCalc,
	}
}

func (uc *CreateMultiStopTrip) Execute(ctx context.Context, input CreateMultiStopTripInput) (*entities.Trip, error) {
	// Calculate base fare using just pickup->dropoff for estimate
	// (Advanced routing through all waypoints can be added later)
	distanceKM := services.CalculateDistanceKM(
		input.PickupLatitude, input.PickupLongitude,
		input.DropoffLatitude, input.DropoffLongitude,
	)
	estimatedMinutes := int(distanceKM * 2)
	estimatedFare := uc.fareCalc.Calculate(distanceKM, estimatedMinutes)

	now := time.Now()
	tripID := uuid.New()

	trip := &entities.Trip{
		ID:               tripID,
		PassengerID:      input.PassengerID,
		TripType:         entities.TripTypeNormal, // or a new MULTI_STOP type if added
		Status:           entities.StatusRequested,
		StartPIN:         "0000", // Simplified for now
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

	// Create waypoints
	if len(input.Waypoints) > 0 {
		var wps []*entities.Waypoint
		for _, wpInput := range input.Waypoints {
			wps = append(wps, &entities.Waypoint{
				ID:        uuid.New(),
				TripID:    tripID,
				Sequence:  wpInput.Sequence,
				Latitude:  wpInput.Latitude,
				Longitude: wpInput.Longitude,
				Address:   wpInput.Address,
				CreatedAt: now,
			})
		}
		if err := uc.waypointRepo.CreateMany(ctx, wps); err != nil {
			return nil, err
		}
		trip.Waypoints = wps // Attach to response
	}

	return trip, nil
}
