package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/geofence/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/geofence/domain/entities"
)

type GeofenceHandler struct {
	createGeofenceUC *usecases.CreateGeofence
	listGeofencesUC  *usecases.ListGeofences
	checkLocationUC  *usecases.CheckLocationInGeofence
	assignDriverUC   *usecases.AssignDriverToZone
	removeDriverUC   *usecases.RemoveDriverFromZone
}

func NewGeofenceHandler(
	createGeofenceUC *usecases.CreateGeofence,
	listGeofencesUC *usecases.ListGeofences,
	checkLocationUC *usecases.CheckLocationInGeofence,
	assignDriverUC *usecases.AssignDriverToZone,
	removeDriverUC *usecases.RemoveDriverFromZone,
) *GeofenceHandler {
	return &GeofenceHandler{
		createGeofenceUC: createGeofenceUC,
		listGeofencesUC:  listGeofencesUC,
		checkLocationUC:  checkLocationUC,
		assignDriverUC:   assignDriverUC,
		removeDriverUC:   removeDriverUC,
	}
}

type CreateGeofenceRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Coordinates string `json:"coordinates" binding:"required"`
	IsActive    bool   `json:"is_active"`
}

func (h *GeofenceHandler) CreateGeofence(c *gin.Context) {
	var req CreateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	fence, err := h.createGeofenceUC.Execute(c.Request.Context(), usecases.CreateGeofenceInput{
		Name:        req.Name,
		Type:        entities.GeofenceType(req.Type),
		Coordinates: req.Coordinates,
		IsActive:    req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_geofence"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"geofence": fence})
}

func (h *GeofenceHandler) ListGeofences(c *gin.Context) {
	activeOnlyStr := c.DefaultQuery("active_only", "true")
	activeOnly, _ := strconv.ParseBool(activeOnlyStr)

	fences, err := h.listGeofencesUC.Execute(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_geofences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofences": fences})
}

func (h *GeofenceHandler) CheckLocation(c *gin.Context) {
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

	zones, err := h.checkLocationUC.Execute(c.Request.Context(), lat, lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_check_location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zones": zones, "in_zone": len(zones) > 0})
}

type AssignDriverRequest struct {
	DriverID string `json:"driver_id" binding:"required"`
	ZoneID   string `json:"zone_id" binding:"required"`
	Status   string `json:"status"`
}

func (h *GeofenceHandler) AssignDriverToZone(c *gin.Context) {
	var req AssignDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	driverID, err := uuid.Parse(req.DriverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_driver_id"})
		return
	}
	zoneID, err := uuid.Parse(req.ZoneID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_zone_id"})
		return
	}

	status := req.Status
	if status == "" {
		status = "WAITING"
	}

	err = h.assignDriverUC.Execute(c.Request.Context(), usecases.AssignDriverToZoneInput{
		DriverID: driverID,
		ZoneID:   zoneID,
		Status:   status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_assign_driver"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "driver_assigned_to_zone"})
}

func (h *GeofenceHandler) RemoveDriverFromZone(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("driverId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_driver_id"})
		return
	}
	zoneID, err := uuid.Parse(c.Param("zoneId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_zone_id"})
		return
	}

	err = h.removeDriverUC.Execute(c.Request.Context(), driverID, zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_remove_driver"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "driver_removed_from_zone"})
}
