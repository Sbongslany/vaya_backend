package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/settings/application/usecases"
)

type AdminSettingsHandler struct {
	createVehicleUC  *usecases.CreateVehicleType
	listVehiclesUC   *usecases.ListVehicleTypes
	updateVehicleUC  *usecases.UpdateVehicleType
	deleteVehicleUC  *usecases.DeleteVehicleType
	getSettingsUC    *usecases.GetPlatformSettings
	updateSettingsUC *usecases.UpdatePlatformSettings
}

func NewAdminSettingsHandler(
	createVehicleUC *usecases.CreateVehicleType,
	listVehiclesUC *usecases.ListVehicleTypes,
	updateVehicleUC *usecases.UpdateVehicleType,
	deleteVehicleUC *usecases.DeleteVehicleType,
	getSettingsUC *usecases.GetPlatformSettings,
	updateSettingsUC *usecases.UpdatePlatformSettings,
) *AdminSettingsHandler {
	return &AdminSettingsHandler{
		createVehicleUC:  createVehicleUC,
		listVehiclesUC:   listVehiclesUC,
		updateVehicleUC:  updateVehicleUC,
		deleteVehicleUC:  deleteVehicleUC,
		getSettingsUC:    getSettingsUC,
		updateSettingsUC: updateSettingsUC,
	}
}

// --- Vehicle Types ---

type CreateVehicleTypeRequest struct {
	Name       string  `json:"name" binding:"required"`
	Slug       string  `json:"slug" binding:"required"`
	BaseFare   float64 `json:"base_fare"`
	PerKmRate  float64 `json:"per_km_rate"`
	PerMinRate float64 `json:"per_min_rate"`
	IsActive   bool    `json:"is_active"`
}

func (h *AdminSettingsHandler) CreateVehicleType(c *gin.Context) {
	var req CreateVehicleTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	v, err := h.createVehicleUC.Execute(c.Request.Context(), usecases.CreateVehicleTypeInput{
		Name:       req.Name,
		Slug:       req.Slug,
		BaseFare:   req.BaseFare,
		PerKmRate:  req.PerKmRate,
		PerMinRate: req.PerMinRate,
		IsActive:   req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_vehicle_type"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"vehicle_type": v})
}

func (h *AdminSettingsHandler) ListVehicleTypes(c *gin.Context) {
	vehicles, err := h.listVehiclesUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_vehicle_types"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vehicle_types": vehicles})
}

type UpdateVehicleTypeRequest struct {
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	BaseFare   float64 `json:"base_fare"`
	PerKmRate  float64 `json:"per_km_rate"`
	PerMinRate float64 `json:"per_min_rate"`
	IsActive   bool    `json:"is_active"`
}

func (h *AdminSettingsHandler) UpdateVehicleType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_vehicle_id"})
		return
	}

	var req UpdateVehicleTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err = h.updateVehicleUC.Execute(c.Request.Context(), usecases.UpdateVehicleTypeInput{
		ID:         id,
		Name:       req.Name,
		Slug:       req.Slug,
		BaseFare:   req.BaseFare,
		PerKmRate:  req.PerKmRate,
		PerMinRate: req.PerMinRate,
		IsActive:   req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_vehicle_type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vehicle_type_updated"})
}

func (h *AdminSettingsHandler) DeleteVehicleType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_vehicle_id"})
		return
	}

	if err := h.deleteVehicleUC.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete_vehicle_type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vehicle_type_deleted"})
}

// --- Platform Settings ---

func (h *AdminSettingsHandler) GetPlatformSettings(c *gin.Context) {
	settings, err := h.getSettingsUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

type UpdatePlatformSettingsRequest struct {
	CommissionPercentage float64 `json:"commission_percentage"`
	CancellationFee      float64 `json:"cancellation_fee"`
}

func (h *AdminSettingsHandler) UpdatePlatformSettings(c *gin.Context) {
	var req UpdatePlatformSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err := h.updateSettingsUC.Execute(c.Request.Context(), usecases.UpdatePlatformSettingsInput{
		CommissionPercentage: req.CommissionPercentage,
		CancellationFee:      req.CancellationFee,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings_updated"})
}
