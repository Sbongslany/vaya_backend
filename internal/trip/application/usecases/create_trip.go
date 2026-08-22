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
	PromoCode        string
}

type CreateTrip struct {
	tripRepo       repositories.TripRepository
	fareCalc       *services.FareCalculator
	eventService   *services.TripEventService
	promoRedeemer  services.PromotionRedeemer
	routingService services.RoutingService
	surgeService   services.SurgeService
}

func NewCreateTrip(
	tripRepo repositories.TripRepository,
	fareCalc *services.FareCalculator,
	eventService *services.TripEventService,
	promoRedeemer services.PromotionRedeemer,
	routingService services.RoutingService,
	surgeService services.SurgeService,
) *CreateTrip {
	return &CreateTrip{
		tripRepo:       tripRepo,
		fareCalc:       fareCalc,
		eventService:   eventService,
		promoRedeemer:  promoRedeemer,
		routingService: routingService,
		surgeService:   surgeService,
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

	// Calculate actual road route
	routeResult, err := uc.routingService.CalculateRoute(
		ctx,
		input.PickupLatitude, input.PickupLongitude,
		input.DropoffLatitude, input.DropoffLongitude,
	)
	if err != nil {
		distanceKM := services.CalculateDistanceKM(
			input.PickupLatitude, input.PickupLongitude,
			input.DropoffLatitude, input.DropoffLongitude,
		)
		estimatedMinutes := int(distanceKM * 2)

		routeResult = &services.RouteResult{
			DistanceKM:      distanceKM,
			DurationMinutes: estimatedMinutes,
		}
	}

	baseFare := uc.fareCalc.Calculate(routeResult.DistanceKM, routeResult.DurationMinutes)

	// Apply surge multiplier
	surgeMultiplier := 1.0
	if uc.surgeService != nil {
		// Record this request as demand
		_ = uc.surgeService.RecordDemand(ctx, input.PickupLatitude, input.PickupLongitude)

		// Get current multiplier for this zone
		multiplier, err := uc.surgeService.GetMultiplier(ctx, input.PickupLatitude, input.PickupLongitude)
		if err == nil && multiplier > 1.0 {
			surgeMultiplier = multiplier
		}
	}

	estimatedFare := baseFare * surgeMultiplier

	now := time.Now()
	tripID := uuid.New()

	trip := &entities.Trip{
		ID:                   tripID,
		PassengerID:          input.PassengerID,
		TripType:             entities.TripTypeNormal,
		Status:               entities.StatusRequested,
		StartPIN:             generatePIN(),
		PickupLatitude:       input.PickupLatitude,
		PickupLongitude:      input.PickupLongitude,
		PickupAddress:        input.PickupAddress,
		DropoffLatitude:      input.DropoffLatitude,
		DropoffLongitude:     input.DropoffLongitude,
		DropoffAddress:       input.DropoffAddress,
		EstimatedFare:        estimatedFare,
		Currency:             "ZAR",
		DistanceKM:           &routeResult.DistanceKM,
		RoutePolyline:        &routeResult.Polyline,
		RouteDurationMinutes: &routeResult.DurationMinutes,
		RouteDistanceKM:      &routeResult.DistanceKM,
		SurgeMultiplier:      &surgeMultiplier,
		DiscountAmount:       0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Apply promo code if provided
	if input.PromoCode != "" && uc.promoRedeemer != nil {
		discount, promoID, err := uc.promoRedeemer.RedeemForTrip(
			ctx, input.PromoCode, input.PassengerID, tripID, estimatedFare,
		)
		if err != nil {
			return nil, err
		}
		trip.PromotionID = &promoID
		trip.DiscountAmount = discount
		trip.EstimatedFare = estimatedFare - discount
	}

	if err := uc.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	// Record trip created event
	_ = uc.eventService.Record(
		ctx,
		trip.ID,
		entities.EventTypeTripCreated,
		&input.PassengerID,
		"",
		string(entities.StatusRequested),
		map[string]interface{}{
			"pickup_address":   input.PickupAddress,
			"dropoff_address":  input.DropoffAddress,
			"estimated_fare":   trip.EstimatedFare,
			"distance_km":      routeResult.DistanceKM,
			"duration_min":     routeResult.DurationMinutes,
			"surge_multiplier": surgeMultiplier,
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
