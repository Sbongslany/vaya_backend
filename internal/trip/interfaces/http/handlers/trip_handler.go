package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type TripHandler struct {
	createTrip                    *usecases.CreateTrip
	getTrip                       *usecases.GetTrip
	getNearbyTrips                *usecases.GetNearbyTrips
	submitTripOffer               *usecases.SubmitTripOffer
	getTripOffers                 *usecases.GetTripOffers
	acceptTripOffer               *usecases.AcceptTripOffer
	confirmTripAssignment         *usecases.ConfirmTripAssignment
	arriveAtPickup                *usecases.ArriveAtPickup
	startTrip                     *usecases.StartTrip
	completeTrip                  *usecases.CompleteTrip
	processPayment                *usecases.ProcessPayment
	submitRating                  *usecases.SubmitRating
	createLongDistanceTrip        *usecases.CreateLongDistanceTrip
	getOpenLongDistanceTrips      *usecases.GetOpenLongDistanceTrips
	publishLongDistanceTrip       *usecases.PublishLongDistanceTrip
	confirmLongDistanceAssignment *usecases.ConfirmLongDistanceAssignment
	scheduleLongDistanceTrip      *usecases.ScheduleLongDistanceTrip
	departForPickup               *usecases.DepartForPickup
	beginOutbound                 *usecases.BeginOutbound
	reachOutboundDestination      *usecases.ReachOutboundDestination
	resolveOutboundArrival        *usecases.ResolveOutboundArrival
	scheduleReturn                *usecases.ScheduleReturn
	startReturn                   *usecases.StartReturn
	beginReturnInProgress         *usecases.BeginReturnInProgress
	reachFinalDestination         *usecases.ReachFinalDestination
	completeLongDistanceTrip      *usecases.CompleteLongDistanceTrip
	cancelTrip                    *usecases.CancelTrip
	calculateRouteUC              *usecases.CalculateRoute // <-- ADD THIS LINE
	getSurgeMultiplierUC          *usecases.GetSurgeMultiplier
	getSurgeHeatmapUC             *usecases.GetSurgeHeatmap
	createMultiStopTripUC         *usecases.CreateMultiStopTrip
}

func NewTripHandler(
	createTrip *usecases.CreateTrip,
	getTrip *usecases.GetTrip,
	getNearbyTrips *usecases.GetNearbyTrips,
	submitTripOffer *usecases.SubmitTripOffer,
	getTripOffers *usecases.GetTripOffers,
	acceptTripOffer *usecases.AcceptTripOffer,
	confirmTripAssignment *usecases.ConfirmTripAssignment,
	arriveAtPickup *usecases.ArriveAtPickup,
	startTrip *usecases.StartTrip,
	completeTrip *usecases.CompleteTrip,
	processPayment *usecases.ProcessPayment,
	submitRating *usecases.SubmitRating,
	createLongDistanceTrip *usecases.CreateLongDistanceTrip,
	getOpenLongDistanceTrips *usecases.GetOpenLongDistanceTrips,
	publishLongDistanceTrip *usecases.PublishLongDistanceTrip,
	confirmLongDistanceAssignment *usecases.ConfirmLongDistanceAssignment,
	scheduleLongDistanceTrip *usecases.ScheduleLongDistanceTrip,
	departForPickup *usecases.DepartForPickup,
	beginOutbound *usecases.BeginOutbound,
	reachOutboundDestination *usecases.ReachOutboundDestination,
	resolveOutboundArrival *usecases.ResolveOutboundArrival,
	scheduleReturn *usecases.ScheduleReturn,
	startReturn *usecases.StartReturn,
	beginReturnInProgress *usecases.BeginReturnInProgress,
	reachFinalDestination *usecases.ReachFinalDestination,
	completeLongDistanceTrip *usecases.CompleteLongDistanceTrip,
	cancelTrip *usecases.CancelTrip,
	calculateRouteUC *usecases.CalculateRoute, // <-- ADD THIS LINE
	getSurgeMultiplierUC *usecases.GetSurgeMultiplier,
	getSurgeHeatmapUC *usecases.GetSurgeHeatmap,
	createMultiStopTripUC *usecases.CreateMultiStopTrip,

) *TripHandler {
	return &TripHandler{
		createTrip:                    createTrip,
		getTrip:                       getTrip,
		getNearbyTrips:                getNearbyTrips,
		submitTripOffer:               submitTripOffer,
		getTripOffers:                 getTripOffers,
		acceptTripOffer:               acceptTripOffer,
		confirmTripAssignment:         confirmTripAssignment,
		arriveAtPickup:                arriveAtPickup,
		startTrip:                     startTrip,
		completeTrip:                  completeTrip,
		processPayment:                processPayment,
		submitRating:                  submitRating,
		createLongDistanceTrip:        createLongDistanceTrip,
		getOpenLongDistanceTrips:      getOpenLongDistanceTrips,
		publishLongDistanceTrip:       publishLongDistanceTrip,
		confirmLongDistanceAssignment: confirmLongDistanceAssignment,
		scheduleLongDistanceTrip:      scheduleLongDistanceTrip,
		departForPickup:               departForPickup,
		beginOutbound:                 beginOutbound,
		reachOutboundDestination:      reachOutboundDestination,
		resolveOutboundArrival:        resolveOutboundArrival,
		scheduleReturn:                scheduleReturn,
		startReturn:                   startReturn,
		beginReturnInProgress:         beginReturnInProgress,
		reachFinalDestination:         reachFinalDestination,
		completeLongDistanceTrip:      completeLongDistanceTrip,
		cancelTrip:                    cancelTrip,
		calculateRouteUC:              calculateRouteUC, // <-- ADD THIS LINE

	}
}

// ==========================================
// NORMAL TRIP HANDLERS
// ==========================================

type CreateTripRequest struct {
	PickupLatitude   float64 `json:"pickup_latitude" binding:"required"`
	PickupLongitude  float64 `json:"pickup_longitude" binding:"required"`
	PickupAddress    string  `json:"pickup_address" binding:"required"`
	DropoffLatitude  float64 `json:"dropoff_latitude" binding:"required"`
	DropoffLongitude float64 `json:"dropoff_longitude" binding:"required"`
	DropoffAddress   string  `json:"dropoff_address" binding:"required"`
}

func (h *TripHandler) CreateTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	trip, err := h.createTrip.Execute(c.Request.Context(), usecases.CreateTripInput{
		PassengerID:      userID,
		PickupLatitude:   req.PickupLatitude,
		PickupLongitude:  req.PickupLongitude,
		PickupAddress:    req.PickupAddress,
		DropoffLatitude:  req.DropoffLatitude,
		DropoffLongitude: req.DropoffLongitude,
		DropoffAddress:   req.DropoffAddress,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"trip": trip})
}

func (h *TripHandler) GetTrip(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.getTrip.Execute(c.Request.Context(), tripID)
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) GetNearbyTrips(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.DefaultQuery("radius", "10")
	limitStr := c.DefaultQuery("limit", "20")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_latitude"})
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_longitude"})
		return
	}
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_radius"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
		return
	}

	trips, err := h.getNearbyTrips.Execute(c.Request.Context(), usecases.GetNearbyTripsInput{
		Latitude:  lat,
		Longitude: lng,
		RadiusKM:  radius,
		Limit:     limit,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trips": trips})
}

// ==========================================
// OFFER HANDLERS
// ==========================================

type SubmitTripOfferRequest struct {
	OfferType   entities.OfferType `json:"offer_type" binding:"required"`
	OfferedFare float64            `json:"offered_fare" binding:"required"`
}

func (h *TripHandler) SubmitTripOffer(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	var req SubmitTripOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	offer, err := h.submitTripOffer.Execute(c.Request.Context(), usecases.SubmitTripOfferInput{
		TripID:      tripID,
		DriverID:    driverID,
		OfferType:   req.OfferType,
		OfferedFare: req.OfferedFare,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"offer": offer})
}

func (h *TripHandler) GetTripOffers(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	offers, err := h.getTripOffers.Execute(c.Request.Context(), tripID)
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"offers": offers})
}

func (h *TripHandler) AcceptTripOffer(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	passengerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	offerID, err := uuid.Parse(c.Param("offerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_offer_id"})
		return
	}

	trip, err := h.acceptTripOffer.Execute(c.Request.Context(), usecases.AcceptTripOfferInput{
		TripID:      tripID,
		OfferID:     offerID,
		PassengerID: passengerID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

// ==========================================
// TRIP EXECUTION HANDLERS (NORMAL)
// ==========================================

func (h *TripHandler) ConfirmTripAssignment(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.confirmTripAssignment.Execute(c.Request.Context(), usecases.ConfirmTripAssignmentInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) ArriveAtPickup(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.arriveAtPickup.Execute(c.Request.Context(), usecases.ArriveAtPickupInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

type StartTripRequest struct {
	PIN string `json:"pin" binding:"required"`
}

func (h *TripHandler) StartTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	var req StartTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	trip, err := h.startTrip.Execute(c.Request.Context(), usecases.StartTripInput{
		TripID:   tripID,
		DriverID: driverID,
		PIN:      req.PIN,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) CompleteTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.completeTrip.Execute(c.Request.Context(), usecases.CompleteTripInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

// ==========================================
// PAYMENT & RATING HANDLERS
// ==========================================

type ProcessPaymentRequest struct {
	Method entities.PaymentMethod `json:"method" binding:"required"`
}

func (h *TripHandler) ProcessPayment(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	passengerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	payment, err := h.processPayment.Execute(c.Request.Context(), usecases.ProcessPaymentInput{
		TripID:      tripID,
		PassengerID: passengerID,
		Method:      req.Method,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

type SubmitRatingRequest struct {
	Rating  int    `json:"rating" binding:"required"`
	Comment string `json:"comment"`
}

func (h *TripHandler) SubmitRating(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	raterID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	var req SubmitRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	rating, err := h.submitRating.Execute(c.Request.Context(), usecases.SubmitRatingInput{
		TripID:  tripID,
		RaterID: raterID,
		Rating:  req.Rating,
		Comment: req.Comment,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"rating": rating})
}

// ==========================================
// LONG-DISTANCE TRIP SETUP HANDLERS
// ==========================================

type CreateLongDistanceTripRequest struct {
	PickupLatitude     float64                   `json:"pickup_latitude" binding:"required"`
	PickupLongitude    float64                   `json:"pickup_longitude" binding:"required"`
	PickupAddress      string                    `json:"pickup_address" binding:"required"`
	DropoffLatitude    float64                   `json:"dropoff_latitude" binding:"required"`
	DropoffLongitude   float64                   `json:"dropoff_longitude" binding:"required"`
	DropoffAddress     string                    `json:"dropoff_address" binding:"required"`
	LongDistanceType   entities.LongDistanceType `json:"long_distance_type" binding:"required"`
	ScheduledDeparture time.Time                 `json:"scheduled_departure" binding:"required"`
	ScheduledReturn    *time.Time                `json:"scheduled_return"`
	TripDurationDays   int                       `json:"trip_duration_days"`
}

func (h *TripHandler) CreateLongDistanceTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	passengerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateLongDistanceTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	trip, err := h.createLongDistanceTrip.Execute(c.Request.Context(), usecases.CreateLongDistanceTripInput{
		PassengerID:        passengerID,
		PickupLatitude:     req.PickupLatitude,
		PickupLongitude:    req.PickupLongitude,
		PickupAddress:      req.PickupAddress,
		DropoffLatitude:    req.DropoffLatitude,
		DropoffLongitude:   req.DropoffLongitude,
		DropoffAddress:     req.DropoffAddress,
		LongDistanceType:   req.LongDistanceType,
		ScheduledDeparture: req.ScheduledDeparture,
		ScheduledReturn:    req.ScheduledReturn,
		TripDurationDays:   req.TripDurationDays,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"trip": trip})
}

func (h *TripHandler) GetOpenLongDistanceTrips(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
		return
	}

	trips, err := h.getOpenLongDistanceTrips.Execute(c.Request.Context(), limit)
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trips": trips})
}

func (h *TripHandler) PublishLongDistanceTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	passengerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.publishLongDistanceTrip.Execute(c.Request.Context(), usecases.PublishLongDistanceTripInput{
		TripID:      tripID,
		PassengerID: passengerID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) ConfirmLongDistanceAssignment(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.confirmLongDistanceAssignment.Execute(c.Request.Context(), usecases.ConfirmLongDistanceAssignmentInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) ScheduleLongDistanceTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	passengerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.scheduleLongDistanceTrip.Execute(c.Request.Context(), usecases.ScheduleLongDistanceTripInput{
		TripID:      tripID,
		PassengerID: passengerID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) DepartForPickup(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.departForPickup.Execute(c.Request.Context(), usecases.DepartForPickupInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

// ==========================================
// LONG-DISTANCE OUTBOUND HANDLERS
// ==========================================

func (h *TripHandler) BeginOutbound(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.beginOutbound.Execute(c.Request.Context(), usecases.BeginOutboundInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) ReachOutboundDestination(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.reachOutboundDestination.Execute(c.Request.Context(), usecases.ReachOutboundDestinationInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) ResolveOutboundArrival(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.resolveOutboundArrival.Execute(c.Request.Context(), usecases.ResolveOutboundArrivalInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

// ==========================================
// LONG-DISTANCE RETURN HANDLERS
// ==========================================

func (h *TripHandler) ScheduleReturn(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.scheduleReturn.Execute(c.Request.Context(), usecases.ScheduleReturnInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) StartReturn(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.startReturn.Execute(c.Request.Context(), usecases.StartReturnInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) BeginReturnInProgress(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.beginReturnInProgress.Execute(c.Request.Context(), usecases.BeginReturnInProgressInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) ReachFinalDestination(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.reachFinalDestination.Execute(c.Request.Context(), usecases.ReachFinalDestinationInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func (h *TripHandler) CompleteLongDistanceTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	trip, err := h.completeLongDistanceTrip.Execute(c.Request.Context(), usecases.CompleteLongDistanceTripInput{
		TripID:   tripID,
		DriverID: driverID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

// ==========================================
// CANCELLATION HANDLER
// ==========================================

type CancelTripRequest struct {
	Reason string `json:"reason"`
}

func (h *TripHandler) CancelTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	var req CancelTripRequest
	c.ShouldBindJSON(&req)

	trip, err := h.cancelTrip.Execute(c.Request.Context(), usecases.CancelTripInput{
		TripID: tripID,
		UserID: userID,
		Reason: req.Reason,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

// ==========================================
// ERROR HANDLER
// ==========================================

func handleTripError(c *gin.Context, err error) {
	switch err {
	case domain.ErrTripNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "trip_not_found"})
	case domain.ErrOfferNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "offer_not_found"})
	case domain.ErrInvalidStateTransition:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_state_transition"})
	case domain.ErrInvalidCoordinates:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_coordinates"})
	case domain.ErrInvalidOfferFare:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_offer_fare"})
	case domain.ErrActiveTripExists:
		c.JSON(http.StatusConflict, gin.H{"error": "active_trip_exists"})
	case domain.ErrUnauthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
	case domain.ErrInvalidPIN:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_pin"})
	case domain.ErrAlreadyPaid:
		c.JSON(http.StatusConflict, gin.H{"error": "already_paid"})
	case domain.ErrInvalidRating:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_rating"})
	case domain.ErrAlreadyRated:
		c.JSON(http.StatusConflict, gin.H{"error": "already_rated"})
	case domain.ErrNotTripParticipant:
		c.JSON(http.StatusForbidden, gin.H{"error": "not_trip_participant"})
	case domain.ErrInvalidSchedule:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_schedule"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	}
}

func (h *TripHandler) GetTripRoute(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	// 👇 CHANGE h.getTripUC to h.getTrip (or whatever your field is named in the struct)
	trip, err := h.getTrip.Execute(c.Request.Context(), tripID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip_not_found"})
		return
	}

	if trip.RoutePolyline == nil {
		c.JSON(http.StatusOK, gin.H{
			"trip_id":  trip.ID,
			"polyline": "",
			"message":  "no_route_calculated_for_this_trip",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trip_id":          trip.ID,
		"polyline":         *trip.RoutePolyline,
		"distance_km":      trip.RouteDistanceKM,
		"duration_minutes": trip.RouteDurationMinutes,
		"pickup":           map[string]float64{"lat": trip.PickupLatitude, "lng": trip.PickupLongitude},
		"dropoff":          map[string]float64{"lat": trip.DropoffLatitude, "lng": trip.DropoffLongitude},
	})
}

type CalculateRouteRequest struct {
	FromLat float64 `json:"from_lat" binding:"required"`
	FromLng float64 `json:"from_lng" binding:"required"`
	ToLat   float64 `json:"to_lat" binding:"required"`
	ToLng   float64 `json:"to_lng" binding:"required"`
}

func (h *TripHandler) CalculateRoute(c *gin.Context) {
	var req CalculateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	result, err := h.calculateRouteUC.Execute(c.Request.Context(), usecases.CalculateRouteInput{
		FromLat: req.FromLat,
		FromLng: req.FromLng,
		ToLat:   req.ToLat,
		ToLng:   req.ToLng,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_calculate_route"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"distance_km":      result.DistanceKM,
		"duration_minutes": result.DurationMinutes,
		"polyline":         result.Polyline,
		"steps":            result.Steps,
	})
}

func (h *TripHandler) GetSurgeMultiplier(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_lat"})
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_lng"})
		return
	}

	multiplier, err := h.getSurgeMultiplierUC.Execute(c.Request.Context(), lat, lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_get_surge"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lat":        lat,
		"lng":        lng,
		"multiplier": multiplier,
	})
}

func (h *TripHandler) GetSurgeHeatmap(c *gin.Context) {
	zones, err := h.getSurgeHeatmapUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_get_heatmap"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

type CreateMultiStopTripRequest struct {
	PickupLatitude   float64                  `json:"pickup_latitude" binding:"required"`
	PickupLongitude  float64                  `json:"pickup_longitude" binding:"required"`
	PickupAddress    string                   `json:"pickup_address" binding:"required"`
	DropoffLatitude  float64                  `json:"dropoff_latitude" binding:"required"`
	DropoffLongitude float64                  `json:"dropoff_longitude" binding:"required"`
	DropoffAddress   string                   `json:"dropoff_address" binding:"required"`
	Waypoints        []usecases.WaypointInput `json:"waypoints"`
}

func (h *TripHandler) CreateMultiStopTrip(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	passengerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateMultiStopTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	trip, err := h.createMultiStopTripUC.Execute(c.Request.Context(), usecases.CreateMultiStopTripInput{
		PassengerID:      passengerID,
		PickupLatitude:   req.PickupLatitude,
		PickupLongitude:  req.PickupLongitude,
		PickupAddress:    req.PickupAddress,
		DropoffLatitude:  req.DropoffLatitude,
		DropoffLongitude: req.DropoffLongitude,
		DropoffAddress:   req.DropoffAddress,
		Waypoints:        req.Waypoints,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"trip": trip})
}
