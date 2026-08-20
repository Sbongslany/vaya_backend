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
	Handler      *handlers.TripHandler
	EventHandler *handlers.TripEventHandler
}

func WireTrip(pgPool *pgxpool.Pool) *TripContainer {
	// Repositories
	tripRepo := postgres.NewTripRepository(pgPool)
	tripOfferRepo := postgres.NewTripOfferRepository(pgPool)
	paymentRepo := postgres.NewPaymentRepository(pgPool)
	ratingRepo := postgres.NewTripRatingRepository(pgPool)
	eventRepo := postgres.NewTripEventRepository(pgPool)

	// Domain services
	fareCalc := services.NewFareCalculator()
	stateMachine := services.NewStateMachine()
	paymentProvider := payment.NewDefaultPaymentProvider()

	// Use cases — normal trip
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

	// Use cases — long-distance trip
	createLongDistanceUC := usecases.NewCreateLongDistanceTrip(tripRepo, fareCalc)
	getOpenLongDistanceUC := usecases.NewGetOpenLongDistanceTrips(tripRepo)
	publishLongDistanceUC := usecases.NewPublishLongDistanceTrip(tripRepo, stateMachine)
	confirmLongDistanceUC := usecases.NewConfirmLongDistanceAssignment(tripRepo, stateMachine)
	scheduleLongDistanceUC := usecases.NewScheduleLongDistanceTrip(tripRepo, stateMachine)
	departForPickupUC := usecases.NewDepartForPickup(tripRepo, stateMachine)
	beginOutboundUC := usecases.NewBeginOutbound(tripRepo, stateMachine)
	reachOutboundUC := usecases.NewReachOutboundDestination(tripRepo, stateMachine)
	resolveOutboundUC := usecases.NewResolveOutboundArrival(tripRepo, stateMachine)
	scheduleReturnUC := usecases.NewScheduleReturn(tripRepo, stateMachine)
	startReturnUC := usecases.NewStartReturn(tripRepo, stateMachine)
	beginReturnUC := usecases.NewBeginReturnInProgress(tripRepo, stateMachine)
	reachFinalUC := usecases.NewReachFinalDestination(tripRepo, stateMachine)
	completeLongDistanceUC := usecases.NewCompleteLongDistanceTrip(tripRepo, stateMachine)

	// Use cases — cancellation
	cancelTripUC := usecases.NewCancelTrip(tripRepo, tripOfferRepo, stateMachine)

	// Use cases — events
	getTripHistoryUC := usecases.NewGetTripHistory(tripRepo, eventRepo)

	// Handlers
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
		scheduleReturnUC,
		startReturnUC,
		beginReturnUC,
		reachFinalUC,
		completeLongDistanceUC,
		cancelTripUC,
	)

	eventHandler := handlers.NewTripEventHandler(getTripHistoryUC)

	return &TripContainer{
		Handler:      handler,
		EventHandler: eventHandler,
	}
}
