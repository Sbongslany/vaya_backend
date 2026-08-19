package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
)

type AdminAuthHandler struct {
	loginUC      *usecases.AdminLogin
	mfaSetupUC   *usecases.AdminMFASetup
	mfaConfirmUC *usecases.AdminMFAConfirm
	mfaVerifyUC  *usecases.AdminMFAVerify
}

func NewAdminAuthHandler(
	login *usecases.AdminLogin,
	mfaSetup *usecases.AdminMFASetup,
	mfaConfirm *usecases.AdminMFAConfirm,
	mfaVerify *usecases.AdminMFAVerify,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		loginUC:      login,
		mfaSetupUC:   mfaSetup,
		mfaConfirmUC: mfaConfirm,
		mfaVerifyUC:  mfaVerify,
	}
}

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	res, err := h.loginUC.Execute(c.Request.Context(), usecases.AdminLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials_or_not_admin"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// MFA Setup generates the TOTP secret and QR code URI
func (h *AdminAuthHandler) MFASetup(c *gin.Context) {
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

	otpAuthURI, err := h.mfaSetupUC.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_setup_mfa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"otpauth_uri": otpAuthURI,
		"message":     "scan_this_uri_with_your_authenticator_app",
	})
}

type AdminMFAConfirmRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// MFA Confirm verifies the first TOTP code to activate MFA
func (h *AdminAuthHandler) MFAConfirm(c *gin.Context) {
	var req AdminMFAConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

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

	err = h.mfaConfirmUC.Execute(c.Request.Context(), userID, req.Code)
	if err != nil {
		if errors.Is(err, domain.ErrMFAInvalidCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_mfa_code"})
			return
		}
		if errors.Is(err, domain.ErrMFAAlreadyEnabled) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mfa_already_enabled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_confirm_mfa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mfa_enabled_successfully"})
}

type AdminMFAVerifyRequest struct {
	MFATicket  string  `json:"mfa_ticket" binding:"required"`
	Code       string  `json:"code" binding:"required,len=6"`
	DeviceID   *string `json:"device_id"`
	DeviceType *string `json:"device_type"`
	DeviceName *string `json:"device_name"`
}

func (h *AdminAuthHandler) MFAVerify(c *gin.Context) {
	var req AdminMFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	res, err := h.mfaVerifyUC.Execute(c.Request.Context(), usecases.AdminMFAVerifyRequest{
		MFATicket:  req.MFATicket,
		Code:       req.Code,
		DeviceID:   req.DeviceID,
		DeviceType: req.DeviceType,
		DeviceName: req.DeviceName,
		IPAddress:  &ip,
		UserAgent:  &ua,
	})

	if err != nil {
		if errors.Is(err, domain.ErrMFAInvalidCode) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_mfa_code"})
			return
		}
		if errors.Is(err, domain.ErrInvalidTokenFormat) || errors.Is(err, domain.ErrMFANotEnabled) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_mfa_ticket"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	c.JSON(http.StatusOK, res)
}