package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
)

type TripEventHandler struct {
	getTripHistory *usecases.GetTripHistory
}

func NewTripEventHandler(getTripHistory *usecases.GetTripHistory) *TripEventHandler {
	return &TripEventHandler{getTripHistory: getTripHistory}
}

func (h *TripEventHandler) GetTripHistory(c *gin.Context) {
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

	events, err := h.getTripHistory.Execute(c.Request.Context(), usecases.GetTripHistoryInput{
		TripID: tripID,
		UserID: userID,
	})
	if err != nil {
		handleTripError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
