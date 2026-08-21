package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/driver/application/usecases"
)

type DriverHandler struct {
	goOnline  *usecases.GoOnline
	goOffline *usecases.GoOffline
	updateLoc *usecases.UpdateLocation
	getNearby *usecases.GetNearbyDrivers
}

func NewDriverHandler(
	goOnline *usecases.GoOnline,
	goOffline *usecases.GoOffline,
	updateLoc *usecases.UpdateLocation,
	getNearby *usecases.GetNearbyDrivers,
) *DriverHandler {
	return &DriverHandler{
		goOnline:  goOnline,
		goOffline: goOffline,
		updateLoc: updateLoc,
		getNearby: getNearby,
	}
}

func (h *DriverHandler) GoOnline(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.goOnline.Execute(c.Request.Context(), userIDStr.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ONLINE"})
}

func (h *DriverHandler) GoOffline(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.goOffline.Execute(c.Request.Context(), userIDStr.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OFFLINE"})
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
}

func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err := h.updateLoc.Execute(c.Request.Context(), usecases.UpdateLocationInput{
		DriverID:  userIDStr.(string),
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Heading:   req.Heading,
		Speed:     req.Speed,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *DriverHandler) GetNearbyDrivers(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.DefaultQuery("radius", "5")

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
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_radius"})
		return
	}

	drivers, err := h.getNearby.Execute(c.Request.Context(), lat, lng, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"drivers": drivers})
}
