package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/safety/application/usecases"
)

type SafetyHandler struct {
	triggerSOSUC     *usecases.TriggerSOS
	generateShareUC  *usecases.GenerateShareLink
	viewSharedTripUC *usecases.ViewSharedTrip
}

func NewSafetyHandler(
	triggerSOSUC *usecases.TriggerSOS,
	generateShareUC *usecases.GenerateShareLink,
	viewSharedTripUC *usecases.ViewSharedTrip,
) *SafetyHandler {
	return &SafetyHandler{
		triggerSOSUC:     triggerSOSUC,
		generateShareUC:  generateShareUC,
		viewSharedTripUC: viewSharedTripUC,
	}
}

type TriggerSOSRequest struct {
	TripID string `json:"trip_id" binding:"required"`
}

func (h *SafetyHandler) TriggerSOS(c *gin.Context) {
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

	var req TriggerSOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	alert, err := h.triggerSOSUC.Execute(c.Request.Context(), tripID, userID)
	if err != nil {
		if err == usecases.ErrTripNotActive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "trip_is_not_active"})
			return
		}
		if err == usecases.ErrSOSAlreadyActive {
			c.JSON(http.StatusConflict, gin.H{"error": "sos_already_active"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_trigger_sos"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"alert": alert, "message": "sos_triggered"})
}

type GenerateShareLinkRequest struct {
	TripID        string `json:"trip_id" binding:"required"`
	DurationHours int    `json:"duration_hours"`
}

func (h *SafetyHandler) GenerateShareLink(c *gin.Context) {
	_, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req GenerateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	shareURL, err := h.generateShareUC.Execute(c.Request.Context(), tripID, req.DurationHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_generate_share_link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"share_url": shareURL})
}

func (h *SafetyHandler) ViewSharedTrip(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_token"})
		return
	}

	view, err := h.viewSharedTripUC.Execute(c.Request.Context(), token)
	if err != nil {
		if err == usecases.ErrInvalidOrExpiredToken {
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid_or_expired_token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_trip"})
		return
	}

	c.JSON(http.StatusOK, view)
}
