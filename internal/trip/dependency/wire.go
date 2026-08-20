package dependency

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/notifications"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/payment"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/websocket"
	"github.com/yourorg/ehailing/backend/internal/trip/interfaces/http/handlers"
)

type TripContainer struct {
	Handler       *handlers.TripHandler
	EventHandler  *handlers.TripEventHandler
	WSHandler     *handlers.WSHandler
	DeviceHandler *handlers.DeviceHandler
}

func WireTrip(pgPool *pgxpool.Pool) *TripContainer {
	// Repositories
	tripRepo := postgres.NewTripRepository(pgPool)
	tripOfferRepo := postgres.NewTripOfferRepository(pgPool)
	paymentRepo := postgres.NewPaymentRepository(pgPool)
	ratingRepo := postgres.NewTripRatingRepository(pgPool)
	eventRepo := postgres.NewTripEventRepository(pgPool)
	deviceTokenRepo := postgres.NewDeviceTokenRepository(pgPool)

	// Domain services
	fareCalc := services.NewFareCalculator()
	stateMachine := services.NewStateMachine()
	paymentProvider := payment.NewDefaultPaymentProvider()

	// WebSocket Hub
	hub := websocket.NewHub()

	// FCM Notification Service
	fcmService, err := notifications.NewFirebaseNotificationService(deviceTokenRepo)
	if err != nil {
		log.Printf("WARNING: Firebase FCM not initialized: %v. Push notifications will be disabled.", err)
		// We pass nil to the event service, which safely skips push notifications if FCM fails to load
		fcmService = nil
	}

	// Event service wired to broadcast to WS AND send FCM pushes
	eventService := services.NewTripEventService(eventRepo, tripRepo, hub, fcmService)

	// Use cases — normal trip
	createTripUC := usecases.NewCreateTrip(tripRepo, fareCalc, eventService)
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

	// Use cases — cancellation, events, & devices
	cancelTripUC := usecases.NewCancelTrip(tripRepo, tripOfferRepo, stateMachine, eventService)
	getTripHistoryUC := usecases.NewGetTripHistory(tripRepo, eventRepo)
	registerDeviceTokenUC := usecases.NewRegisterDeviceToken(deviceTokenRepo)

	// Handlers
	handler := handlers.NewTripHandler(
		createTripUC, getTripUC, getNearbyTripsUC, submitOfferUC, getOffersUC,
		acceptOfferUC, confirmAssignmentUC, arriveAtPickupUC, startTripUC, completeTripUC,
		processPaymentUC, submitRatingUC, createLongDistanceUC, getOpenLongDistanceUC,
		publishLongDistanceUC, confirmLongDistanceUC, scheduleLongDistanceUC, departForPickupUC,
		beginOutboundUC, reachOutboundUC, resolveOutboundUC, scheduleReturnUC, startReturnUC,
		beginReturnUC, reachFinalUC, completeLongDistanceUC, cancelTripUC,
	)

	eventHandler := handlers.NewTripEventHandler(getTripHistoryUC)
	wsHandler := handlers.NewWSHandler(hub)
	deviceHandler := handlers.NewDeviceHandler(registerDeviceTokenUC)

	return &TripContainer{
		Handler:       handler,
		EventHandler:  eventHandler,
		WSHandler:     wsHandler,
		DeviceHandler: deviceHandler,
	}
}
