package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/payment"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/trip/interfaces/http/handlers"
)

type TripContainer struct {
	Handler *handlers.TripHandler
}

func WireTrip(pgPool *pgxpool.Pool) *TripContainer {
	// Repositories
	tripRepo := postgres.NewTripRepository(pgPool)
	tripOfferRepo := postgres.NewTripOfferRepository(pgPool)
	paymentRepo := postgres.NewPaymentRepository(pgPool)
	ratingRepo := postgres.NewTripRatingRepository(pgPool)

	// Domain services
	fareCalc := services.NewFareCalculator()
	stateMachine := services.NewStateMachine()
	paymentProvider := payment.NewDefaultPaymentProvider()

	// Use cases
	createTripUC := usecases.NewCreateTrip(tripRepo, fareCalc)
	getTripUC := usecases.NewGetTrip(tripRepo)
	getNearbyTripsUC := usecases.NewGetNearbyTrips(tripRepo)
	submitOfferUC := usecases.NewSubmitTripOffer(tripRepo, tripOfferRepo, stateMachine)
	getOffersUC := usecases.NewGetTripOffers(tripRepo, tripOfferRepo)
	acceptOfferUC := usecases.NewAcceptTripOffer(tripRepo, tripOfferRepo, stateMachine)
	confirmAssignmentUC := usecases.NewConfirmTripAssignment(tripRepo, stateMachine)
	arriveAtPickupUC := usecases.NewArriveAtPickup(tripRepo, stateMachine)
	startTripUC := usecases.NewStartTrip(tripRepo, stateMachine)
	completeTripUC := usecases.NewCompleteTrip(tripRepo, stateMachine)
	processPaymentUC := usecases.NewProcessPayment(tripRepo, paymentRepo, paymentProvider, stateMachine)
	submitRatingUC := usecases.NewSubmitRating(tripRepo, ratingRepo, stateMachine)
	createLongDistanceUC := usecases.NewCreateLongDistanceTrip(tripRepo, fareCalc)
	getOpenLongDistanceUC := usecases.NewGetOpenLongDistanceTrips(tripRepo)
	publishLongDistanceUC := usecases.NewPublishLongDistanceTrip(tripRepo, stateMachine)
	confirmLongDistanceUC := usecases.NewConfirmLongDistanceAssignment(tripRepo, stateMachine)
	scheduleLongDistanceUC := usecases.NewScheduleLongDistanceTrip(tripRepo, stateMachine)
	departForPickupUC := usecases.NewDepartForPickup(tripRepo, stateMachine)
	beginOutboundUC := usecases.NewBeginOutbound(tripRepo, stateMachine)
	reachOutboundUC := usecases.NewReachOutboundDestination(tripRepo, stateMachine)
	resolveOutboundUC := usecases.NewResolveOutboundArrival(tripRepo, stateMachine)

	// Handler
	handler := handlers.NewTripHandler(
		createTripUC,
		getTripUC,
		getNearbyTripsUC,
		submitOfferUC,
		getOffersUC,
		acceptOfferUC,
		confirmAssignmentUC,
		arriveAtPickupUC,
		startTripUC,
		completeTripUC,
		processPaymentUC,
		submitRatingUC,
		createLongDistanceUC,
		getOpenLongDistanceUC,
		publishLongDistanceUC,
		confirmLongDistanceUC,
		scheduleLongDistanceUC,
		departForPickupUC,
		beginOutboundUC,
		reachOutboundUC,
		resolveOutboundUC,
	)

	return &TripContainer{
		Handler: handler,
	}
}
