package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type TripHandler struct {
	createTrip            *usecases.CreateTrip
	getTrip               *usecases.GetTrip
	getNearbyTrips        *usecases.GetNearbyTrips
	submitTripOffer       *usecases.SubmitTripOffer
	getTripOffers         *usecases.GetTripOffers
	acceptTripOffer       *usecases.AcceptTripOffer
	confirmTripAssignment *usecases.ConfirmTripAssignment
	arriveAtPickup        *usecases.ArriveAtPickup
	startTrip             *usecases.StartTrip
	completeTrip          *usecases.CompleteTrip
	processPayment        *usecases.ProcessPayment
	submitRating          *usecases.SubmitRating
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
) *TripHandler {
	return &TripHandler{
		createTrip:            createTrip,
		getTrip:               getTrip,
		getNearbyTrips:        getNearbyTrips,
		submitTripOffer:       submitTripOffer,
		getTripOffers:         getTripOffers,
		acceptTripOffer:       acceptTripOffer,
		confirmTripAssignment: confirmTripAssignment,
		arriveAtPickup:        arriveAtPickup,
		startTrip:             startTrip,
		completeTrip:          completeTrip,
		processPayment:        processPayment,
		submitRating:          submitRating,
	}
}

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
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	}
}
