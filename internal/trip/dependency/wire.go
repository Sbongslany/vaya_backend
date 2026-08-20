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
	eventService := services.NewTripEventService(eventRepo)

	// Use cases — normal trip
	createTripUC := usecases.NewCreateTrip(tripRepo, fareCalc, eventService)
	getTripUC := usecases.NewGetTrip(tripRepo)
	getNearbyTripsUC := usecases.NewGetNearbyTrips(tripRepo)
	submitTripOfferUC := usecases.NewSubmitTripOffer(tripRepo, tripOfferRepo, stateMachine)
	getTripOffersUC := usecases.NewGetTripOffers(tripRepo, tripOfferRepo)
	acceptTripOfferUC := usecases.NewAcceptTripOffer(tripRepo, tripOfferRepo, stateMachine)
	confirmTripAssignmentUC := usecases.NewConfirmTripAssignment(tripRepo, stateMachine)
	arriveAtPickupUC := usecases.NewArriveAtPickup(tripRepo, stateMachine)
	startTripUC := usecases.NewStartTrip(tripRepo, stateMachine)
	completeTripUC := usecases.NewCompleteTrip(tripRepo, stateMachine)
	processPaymentUC := usecases.NewProcessPayment(tripRepo, paymentRepo, paymentProvider, stateMachine)
	submitRatingUC := usecases.NewSubmitRating(tripRepo, ratingRepo, stateMachine)

	// Use cases — long-distance trip
	createLongDistanceTripUC := usecases.NewCreateLongDistanceTrip(tripRepo, fareCalc)
	getOpenLongDistanceTripsUC := usecases.NewGetOpenLongDistanceTrips(tripRepo)
	publishLongDistanceTripUC := usecases.NewPublishLongDistanceTrip(tripRepo, stateMachine)
	confirmLongDistanceAssignmentUC := usecases.NewConfirmLongDistanceAssignment(tripRepo, stateMachine)
	scheduleLongDistanceTripUC := usecases.NewScheduleLongDistanceTrip(tripRepo, stateMachine)
	departForPickupUC := usecases.NewDepartForPickup(tripRepo, stateMachine)
	beginOutboundUC := usecases.NewBeginOutbound(tripRepo, stateMachine)
	reachOutboundDestinationUC := usecases.NewReachOutboundDestination(tripRepo, stateMachine)
	resolveOutboundArrivalUC := usecases.NewResolveOutboundArrival(tripRepo, stateMachine)
	scheduleReturnUC := usecases.NewScheduleReturn(tripRepo, stateMachine)
	startReturnUC := usecases.NewStartReturn(tripRepo, stateMachine)
	beginReturnInProgressUC := usecases.NewBeginReturnInProgress(tripRepo, stateMachine)
	reachFinalDestinationUC := usecases.NewReachFinalDestination(tripRepo, stateMachine)
	completeLongDistanceTripUC := usecases.NewCompleteLongDistanceTrip(tripRepo, stateMachine)

	// Use cases — cancellation & events
	cancelTripUC := usecases.NewCancelTrip(tripRepo, tripOfferRepo, stateMachine, eventService)
	getTripHistoryUC := usecases.NewGetTripHistory(tripRepo, eventRepo)

	// Handlers
	handler := handlers.NewTripHandler(
		createTripUC,
		getTripUC,
		getNearbyTripsUC,
		submitTripOfferUC,
		getTripOffersUC,
		acceptTripOfferUC,
		confirmTripAssignmentUC,
		arriveAtPickupUC,
		startTripUC,
		completeTripUC,
		processPaymentUC,
		submitRatingUC,
		createLongDistanceTripUC,
		getOpenLongDistanceTripsUC,
		publishLongDistanceTripUC,
		confirmLongDistanceAssignmentUC,
		scheduleLongDistanceTripUC,
		departForPickupUC,
		beginOutboundUC,
		reachOutboundDestinationUC,
		resolveOutboundArrivalUC,
		scheduleReturnUC,
		startReturnUC,
		beginReturnInProgressUC,
		reachFinalDestinationUC,
		completeLongDistanceTripUC,
		cancelTripUC,
	)

	eventHandler := handlers.NewTripEventHandler(getTripHistoryUC)

	return &TripContainer{
		Handler:      handler,
		EventHandler: eventHandler,
	}
}
