package dependency

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	driverRedis "github.com/yourorg/ehailing/backend/internal/driver/infrastructure/persistence/redis"
	promoDep "github.com/yourorg/ehailing/backend/internal/promotions/dependency"
	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/notifications"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/payment"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/websocket"
	"github.com/yourorg/ehailing/backend/internal/trip/interfaces/http/handlers"
	walletDep "github.com/yourorg/ehailing/backend/internal/wallet/dependency" // <-- ADDED
)

type TripContainer struct {
	Handler       *handlers.TripHandler
	EventHandler  *handlers.TripEventHandler
	WSHandler     *handlers.WSHandler
	DeviceHandler *handlers.DeviceHandler
	RatingHandler *handlers.RatingHandler
}

// UPDATED SIGNATURE: Added walletContainer parameter
func WireTrip(pgPool *pgxpool.Pool, promosContainer *promoDep.PromotionsContainer, redisClient *redis.Client, walletContainer *walletDep.WalletContainer) *TripContainer {
	// Repositories
	tripRepo := postgres.NewTripRepository(pgPool)
	tripOfferRepo := postgres.NewTripOfferRepository(pgPool)
	paymentRepo := postgres.NewPaymentRepository(pgPool)
	ratingRepo := postgres.NewTripRatingRepository(pgPool)
	eventRepo := postgres.NewTripEventRepository(pgPool)
	deviceTokenRepo := postgres.NewDeviceTokenRepository(pgPool)
	userRatingRepo := postgres.NewUserRatingRepository(pgPool)

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
		fcmService = nil
	}

	// Event service wired to broadcast to WS AND send FCM pushes
	eventService := services.NewTripEventService(eventRepo, tripRepo, hub, fcmService)

	// Driver state manager (marks drivers BUSY/ONLINE)
	driverStateRepo := driverRedis.NewDriverStateRepository(redisClient)
	driverStateManager := driverRedis.NewDriverStateManagerAdapter(driverStateRepo)

	// Promotion redeemer adapter
	promoRedeemer := promoDep.NewPromotionRedeemerAdapter(promosContainer.RedeemUC)

	// Fare splitter adapter (connects to Wallet module)
	var fareSplitter services.FareSplitter
	if walletContainer != nil && walletContainer.SplitTripFare != nil {
		fareSplitter = walletDep.NewFareSplitterAdapter(walletContainer.SplitTripFare)
	}

	// Use cases — normal trip
	createTripUC := usecases.NewCreateTrip(tripRepo, fareCalc, eventService, promoRedeemer)
	getTripUC := usecases.NewGetTrip(tripRepo)
	getNearbyTripsUC := usecases.NewGetNearbyTrips(tripRepo)
	submitOfferUC := usecases.NewSubmitTripOffer(tripRepo, tripOfferRepo, stateMachine)
	getOffersUC := usecases.NewGetTripOffers(tripRepo, tripOfferRepo)
	acceptOfferUC := usecases.NewAcceptTripOffer(tripRepo, tripOfferRepo, stateMachine, driverStateManager)
	confirmAssignmentUC := usecases.NewConfirmTripAssignment(tripRepo, stateMachine)
	arriveAtPickupUC := usecases.NewArriveAtPickup(tripRepo, stateMachine)
	startTripUC := usecases.NewStartTrip(tripRepo, stateMachine)

	// UPDATED: Pass fareSplitter
	completeTripUC := usecases.NewCompleteTrip(tripRepo, stateMachine, driverStateManager, fareSplitter)

	processPaymentUC := usecases.NewProcessPayment(tripRepo, paymentRepo, paymentProvider, stateMachine)
	submitRatingUC := usecases.NewSubmitRating(tripRepo, ratingRepo, userRatingRepo, stateMachine, eventService)

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

	// UPDATED: Pass fareSplitter
	completeLongDistanceUC := usecases.NewCompleteLongDistanceTrip(tripRepo, stateMachine, fareSplitter)

	// Use cases — cancellation, events, devices & ratings
	cancelTripUC := usecases.NewCancelTrip(tripRepo, tripOfferRepo, stateMachine, eventService, driverStateManager)
	getTripHistoryUC := usecases.NewGetTripHistory(tripRepo, eventRepo)
	registerDeviceTokenUC := usecases.NewRegisterDeviceToken(deviceTokenRepo)
	getUserRatingUC := usecases.NewGetUserRating(userRatingRepo)

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
	ratingHandler := handlers.NewRatingHandler(getUserRatingUC)

	return &TripContainer{
		Handler:       handler,
		EventHandler:  eventHandler,
		WSHandler:     wsHandler,
		DeviceHandler: deviceHandler,
		RatingHandler: ratingHandler,
	}
}
