package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
)

type DriverHandler struct {
	initOnboardingUC    *usecases.InitiateDriverOnboarding
	updateProfileUC     *usecases.UpdateDriverProfile
	createVehicleUC     *usecases.CreateVehicle
	getOnboardingUC     *usecases.GetOnboardingStatus
	generateSignatureUC *usecases.GenerateUploadSignature // NEW
	submitDocumentUC    *usecases.SubmitDocument          // NEW
}

func NewDriverHandler(
	initOnboarding *usecases.InitiateDriverOnboarding,
	updateProfile *usecases.UpdateDriverProfile,
	createVehicle *usecases.CreateVehicle,
	getOnboarding *usecases.GetOnboardingStatus,
	generateSignature *usecases.GenerateUploadSignature, // NEW
	submitDocument *usecases.SubmitDocument, // NEW
) *DriverHandler {
	return &DriverHandler{
		initOnboardingUC: initOnboarding,
		updateProfileUC:  updateProfile,
		createVehicleUC:  createVehicle,
		getOnboardingUC:  getOnboarding,
	}
}

// InitOnboarding starts the driver onboarding process
func (h *DriverHandler) InitOnboarding(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	profile, err := h.initOnboardingUC.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_initiate_onboarding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "driver_onboarding_initiated",
		"profile": profile,
	})
}

type UpdateDriverProfileRequest struct {
	LicenseNumber *string    `json:"license_number"`
	LicenseExpiry *time.Time `json:"license_expiry"`
}

func (h *DriverHandler) UpdateProfile(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	var req UpdateDriverProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	// Fetch the driver profile for this user
	status, err := h.getOnboardingUC.Execute(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "driver_profile_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_profile"})
		return
	}

	err = h.updateProfileUC.Execute(c.Request.Context(), usecases.UpdateDriverProfileRequest{
		ProfileID:     status.Profile.ID,
		LicenseNumber: req.LicenseNumber,
		LicenseExpiry: req.LicenseExpiry,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile_updated_successfully"})
}

type CreateVehicleRequest struct {
	Make        string             `json:"make" binding:"required"`
	Model       string             `json:"model" binding:"required"`
	Year        int                `json:"year" binding:"required"`
	Color       string             `json:"color" binding:"required"`
	PlateNumber string             `json:"plate_number" binding:"required"`
	VehicleType domain.VehicleType `json:"vehicle_type" binding:"required"`
}

func (h *DriverHandler) CreateVehicle(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	var req CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	// Fetch the driver profile for this user
	status, err := h.getOnboardingUC.Execute(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "driver_profile_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_profile"})
		return
	}

	vehicle, err := h.createVehicleUC.Execute(c.Request.Context(), usecases.CreateVehicleRequest{
		DriverProfileID: status.Profile.ID,
		Make:            req.Make,
		Model:           req.Model,
		Year:            req.Year,
		Color:           req.Color,
		PlateNumber:     req.PlateNumber,
		VehicleType:     req.VehicleType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_vehicle"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "vehicle_created_successfully",
		"vehicle": vehicle,
	})
}

// GetOnboardingStatus returns the current onboarding state for the Flutter app
func (h *DriverHandler) GetOnboardingStatus(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	status, err := h.getOnboardingUC.Execute(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "driver_profile_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_onboarding_status"})
		return
	}

	c.JSON(http.StatusOK, status)
}
func (h *DriverHandler) GetUploadSignature(c *gin.Context) {
	userIDStr, _ := c.Get(authMiddleware.UserIDKey)
	userID, _ := uuid.Parse(userIDStr.(string))

	docType := c.Query("doc_type")
	if docType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doc_type query parameter is required"})
		return
	}

	signature, err := h.generateSignatureUC.Execute(c.Request.Context(), usecases.GenerateUploadSignatureRequest{
		UserID:  userID,
		DocType: docType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_generate_signature"})
		return
	}

	c.JSON(http.StatusOK, signature)
}

type SubmitDocumentRequest struct {
	VehicleID *uuid.UUID          `json:"vehicle_id"`
	DocType   domain.DocumentType `json:"doc_type" binding:"required"`
	FileKey   string              `json:"file_key" binding:"required"`
	FileURL   string              `json:"file_url" binding:"required"`
}

func (h *DriverHandler) SubmitDocument(c *gin.Context) {
	userIDStr, _ := c.Get(authMiddleware.UserIDKey)
	userID, _ := uuid.Parse(userIDStr.(string))

	var req SubmitDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err := h.submitDocumentUC.Execute(c.Request.Context(), usecases.SubmitDocumentRequest{
		UserID:    userID,
		VehicleID: req.VehicleID,
		DocType:   req.DocType,
		FileKey:   req.FileKey,
		FileURL:   req.FileURL,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_submit_document"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "document_submitted_successfully"})
}
