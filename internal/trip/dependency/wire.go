package dependency

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	driverRedis "github.com/yourorg/ehailing/backend/internal/driver/infrastructure/persistence/redis"
	promoDep "github.com/yourorg/ehailing/backend/internal/promotions/dependency"
	settingsPostgres "github.com/yourorg/ehailing/backend/internal/settings/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/notifications"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/payment"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/routing"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/surge"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/websocket"
	"github.com/yourorg/ehailing/backend/internal/trip/interfaces/http/handlers"
	walletUseCases "github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
	walletDep "github.com/yourorg/ehailing/backend/internal/wallet/dependency"
	walletPostgres "github.com/yourorg/ehailing/backend/internal/wallet/infrastructure/persistence/postgres"
)

type TripContainer struct {
	Handler        *handlers.TripHandler
	EventHandler   *handlers.TripEventHandler
	WSHandler      *handlers.WSHandler
	DeviceHandler  *handlers.DeviceHandler
	RatingHandler  *handlers.RatingHandler
	PaymentHandler *handlers.PaymentHandler
}

func WireTrip(
	pgPool *pgxpool.Pool,
	promosContainer *promoDep.PromotionsContainer,
	redisClient *redis.Client,
	walletContainer *walletDep.WalletContainer,
	paystackSecretKey string,
	paystackCallbackURL string,
	routingProvider string,
	osrmBaseURL string,
	googleMapsAPIKey string,
) *TripContainer {
	// Repositories
	tripRepo := postgres.NewTripRepository(pgPool)
	waypointRepo := postgres.NewWaypointRepository(pgPool)
	tripOfferRepo := postgres.NewTripOfferRepository(pgPool)
	paymentRepo := postgres.NewPaymentRepository(pgPool)
	ratingRepo := postgres.NewTripRatingRepository(pgPool)
	eventRepo := postgres.NewTripEventRepository(pgPool)
	deviceTokenRepo := postgres.NewDeviceTokenRepository(pgPool)
	userRatingRepo := postgres.NewUserRatingRepository(pgPool)

	// Domain services
	fareCalc := services.NewFareCalculator()
	stateMachine := services.NewStateMachine()

	// Routing service
	var routingService services.RoutingService
	switch routingProvider {
	case "osrm":
		routingService = routing.NewOSRMClient(osrmBaseURL)
		log.Println("Using OSRM routing service")
	case "google":
		routingService = routing.NewGoogleMapsClient(googleMapsAPIKey)
		log.Println("Using Google Maps routing service")
	default:
		routingService = routing.NewHaversineFallback()
		log.Println("Using Haversine fallback routing")
	}

	// Surge pricing service
	surgeService := surge.NewRedisSurgeService(redisClient)

	// Paystack service
	paystackService := payment.NewPaystackService(paystackSecretKey)

	// WebSocket Hub
	hub := websocket.NewHub()

	// FCM Notification Service
	fcmService, err := notifications.NewFirebaseNotificationService(deviceTokenRepo)
	if err != nil {
		log.Printf("WARNING: Firebase FCM not initialized: %v", err)
		fcmService = nil
	}

	// Event service
	eventService := services.NewTripEventService(eventRepo, tripRepo, hub, fcmService)

	// Driver state manager
	driverStateRepo := driverRedis.NewDriverStateRepository(redisClient)
	driverStateManager := driverRedis.NewDriverStateManagerAdapter(driverStateRepo)

	// Promotion redeemer adapter
	promoRedeemer := promoDep.NewPromotionRedeemerAdapter(promosContainer.RedeemUC)

	// Fare splitter adapter
	var fareSplitter services.FareSplitter
	if walletContainer != nil && walletContainer.SplitTripFare != nil {
		fareSplitter = walletDep.NewFareSplitterAdapter(walletContainer.SplitTripFare)
	}

	// Wallet balance use case
	walletPostgresRepo := walletPostgres.NewWalletRepository(pgPool)
	walletBalanceUC := walletUseCases.NewGetWallet(walletPostgresRepo)
	vehicleTypeRepo := settingsPostgres.NewVehicleTypeRepository(pgPool)

	// Use cases — normal trip
	createTripUC := usecases.NewCreateTrip(tripRepo, vehicleTypeRepo, fareCalc, eventService, promoRedeemer, routingService, surgeService)
	createMultiStopTripUC := usecases.NewCreateMultiStopTrip(tripRepo, waypointRepo, fareCalc)
	getTripUC := usecases.NewGetTrip(tripRepo)
	getNearbyTripsUC := usecases.NewGetNearbyTrips(tripRepo)
	submitOfferUC := usecases.NewSubmitTripOffer(tripRepo, tripOfferRepo, stateMachine)
	getOffersUC := usecases.NewGetTripOffers(tripRepo, tripOfferRepo)
	acceptOfferUC := usecases.NewAcceptTripOffer(tripRepo, tripOfferRepo, stateMachine, driverStateManager)
	confirmAssignmentUC := usecases.NewConfirmTripAssignment(tripRepo, stateMachine)
	arriveAtPickupUC := usecases.NewArriveAtPickup(tripRepo, stateMachine)
	startTripUC := usecases.NewStartTrip(tripRepo, stateMachine)
	completeTripUC := usecases.NewCompleteTrip(tripRepo, stateMachine, driverStateManager, fareSplitter)
	paymentProvider := payment.NewDefaultPaymentProvider()
	processPaymentUC := usecases.NewProcessPayment(tripRepo, paymentRepo, paymentProvider, stateMachine)
	submitRatingUC := usecases.NewSubmitRating(tripRepo, ratingRepo, userRatingRepo, stateMachine, eventService)
	calculateRouteUC := usecases.NewCalculateRoute(routingService)
	getSurgeMultiplierUC := usecases.NewGetSurgeMultiplier(surgeService)
	getSurgeHeatmapUC := usecases.NewGetSurgeHeatmap(surgeService)

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
	completeLongDistanceUC := usecases.NewCompleteLongDistanceTrip(tripRepo, stateMachine, fareSplitter)

	// Use cases — cancellation, events, devices & ratings
	cancelTripUC := usecases.NewCancelTrip(tripRepo, tripOfferRepo, stateMachine, eventService, driverStateManager)
	getTripHistoryUC := usecases.NewGetTripHistory(tripRepo, eventRepo)
	registerDeviceTokenUC := usecases.NewRegisterDeviceToken(deviceTokenRepo)
	getUserRatingUC := usecases.NewGetUserRating(userRatingRepo)

	// Use cases — payments
	initiatePaymentUC := usecases.NewInitiateTripPayment(
		tripRepo, paymentRepo, paystackService, walletBalanceUC, paystackCallbackURL,
	)
	handleWebhookUC := usecases.NewHandlePaystackWebhook(
		paymentRepo, tripRepo, paystackService, stateMachine,
	)

	// Handlers
	handler := handlers.NewTripHandler(
		createTripUC, getTripUC, getNearbyTripsUC, submitOfferUC, getOffersUC,
		acceptOfferUC, confirmAssignmentUC, arriveAtPickupUC, startTripUC, completeTripUC,
		processPaymentUC, submitRatingUC, createLongDistanceUC, getOpenLongDistanceUC,
		publishLongDistanceUC, confirmLongDistanceUC, scheduleLongDistanceUC, departForPickupUC,
		beginOutboundUC, reachOutboundUC, resolveOutboundUC, scheduleReturnUC, startReturnUC,
		beginReturnUC, reachFinalUC, completeLongDistanceUC, cancelTripUC, calculateRouteUC,
		getSurgeMultiplierUC, getSurgeHeatmapUC, createMultiStopTripUC,
	)

	eventHandler := handlers.NewTripEventHandler(getTripHistoryUC)
	wsHandler := handlers.NewWSHandler(hub)
	deviceHandler := handlers.NewDeviceHandler(registerDeviceTokenUC)
	ratingHandler := handlers.NewRatingHandler(getUserRatingUC)
	paymentHandler := handlers.NewPaymentHandler(initiatePaymentUC, handleWebhookUC, paystackService)

	return &TripContainer{
		Handler:        handler,
		EventHandler:   eventHandler,
		WSHandler:      wsHandler,
		DeviceHandler:  deviceHandler,
		RatingHandler:  ratingHandler,
		PaymentHandler: paymentHandler,
	}
}
