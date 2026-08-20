package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
)

type DeviceHandler struct {
	registerToken *usecases.RegisterDeviceToken
}

func NewDeviceHandler(registerToken *usecases.RegisterDeviceToken) *DeviceHandler {
	return &DeviceHandler{registerToken: registerToken}
}

type RegisterTokenRequest struct {
	Token      string `json:"token" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"` // "IOS" or "ANDROID"
}

func (h *DeviceHandler) RegisterToken(c *gin.Context) {
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

	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	err = h.registerToken.Execute(c.Request.Context(), usecases.RegisterDeviceTokenInput{
		UserID:     userID,
		Token:      req.Token,
		DeviceType: req.DeviceType,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_register_token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token_registered"})
}
