package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/payment"
)

type PaymentHandler struct {
	initiatePayment *usecases.InitiateTripPayment
	handleWebhook   *usecases.HandlePaystackWebhook
	paystackService *payment.PaystackService
}

func NewPaymentHandler(
	initiatePayment *usecases.InitiateTripPayment,
	handleWebhook *usecases.HandlePaystackWebhook,
	paystackService *payment.PaystackService,
) *PaymentHandler {
	return &PaymentHandler{
		initiatePayment: initiatePayment,
		handleWebhook:   handleWebhook,
		paystackService: paystackService,
	}
}

type InitiatePaymentRequest struct {
	Method string `json:"method" binding:"required"` // CASH, CARD, or WALLET
}

func (h *PaymentHandler) InitiatePayment(c *gin.Context) {
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

	var req InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	// Get passenger email from context (set by auth middleware)
	passengerEmail := ""
	if emailStr, emailExists := c.Get("user_email"); emailExists {
		passengerEmail = emailStr.(string)
	}

	result, err := h.initiatePayment.Execute(c.Request.Context(), usecases.InitiateTripPaymentInput{
		TripID:         tripID,
		PassengerID:    passengerID,
		Method:         entities.PaymentMethod(req.Method),
		PassengerEmail: passengerEmail,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_id":        result.PaymentID,
		"method":            result.Method,
		"status":            result.Status,
		"authorization_url": result.AuthorizationURL,
	})
}

// HandlePaystackWebhook receives webhook events from Paystack
// This endpoint is PUBLIC (no auth) because Paystack calls it directly
func (h *PaymentHandler) HandlePaystackWebhook(c *gin.Context) {
	// Read the raw body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify the signature
	signature := c.GetHeader("x-paystack-signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature"})
		return
	}

	if !h.paystackService.VerifyWebhookSignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Parse the webhook event
	var event struct {
		Event string `json:"event"`
		Data  struct {
			Reference string `json:"reference"`
			Status    string `json:"status"`
		} `json:"data"`
	}

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	// Only process charge.success events
	if event.Event != "charge.success" {
		c.JSON(http.StatusOK, gin.H{"message": "event ignored"})
		return
	}

	// Process the payment
	if err := h.handleWebhook.Execute(c.Request.Context(), event.Data.Reference, event.Data.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment processed"})
}

func handlePaymentError(c *gin.Context, err error) {
	switch err {
	case domain.ErrTripNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "trip_not_found"})
	case domain.ErrAlreadyPaid:
		c.JSON(http.StatusConflict, gin.H{"error": "already_paid"})
	case domain.ErrInvalidStateTransition:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_state_transition"})
	case domain.ErrUnauthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	}
}
