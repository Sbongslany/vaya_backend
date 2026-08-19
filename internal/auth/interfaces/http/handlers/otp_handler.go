package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/ehailing/backend/internal/auth/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
)

type OTPHandler struct {
	requestUC *usecases.RequestOTP
	verifyUC  *usecases.VerifyOTP
}

func NewOTPHandler(request *usecases.RequestOTP, verify *usecases.VerifyOTP) *OTPHandler {
	return &OTPHandler{requestUC: request, verifyUC: verify}
}

type RequestOTPRequest struct {
	Identifier string            `json:"identifier" binding:"required"`
	Purpose    domain.OTPPurpose `json:"purpose" binding:"required"`
	Channel    domain.OTPChannel `json:"channel" binding:"required"`
}

func (h *OTPHandler) Request(c *gin.Context) {
	var req RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	if err := h.requestUC.Execute(c.Request.Context(), req.Identifier, req.Purpose, req.Channel); err != nil {
		if errors.Is(err, domain.ErrOTPCooldownActive) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "otp_cooldown_active"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_send_otp"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "otp_sent_successfully"})
}

type VerifyOTPRequest struct {
	Identifier string            `json:"identifier" binding:"required"`
	Purpose    domain.OTPPurpose `json:"purpose" binding:"required"`
	OTP        string            `json:"otp" binding:"required"`
}

func (h *OTPHandler) Verify(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	if err := h.verifyUC.Execute(c.Request.Context(), req.Identifier, req.Purpose, req.OTP); err != nil {
		if errors.Is(err, domain.ErrOTPInvalid) || errors.Is(err, domain.ErrOTPNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_or_expired_otp"})
			return
		}
		if errors.Is(err, domain.ErrOTPMaxAttemptsExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "max_otp_attempts_exceeded"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_verify_otp"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "otp_verified_successfully"})
}
