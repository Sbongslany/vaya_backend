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

type VerificationHandler struct {
	verifyPhoneUC      *usecases.VerifyPhone
	requestEmailUC     *usecases.RequestEmailVerification
	verifyEmailTokenUC *usecases.VerifyEmailToken
}

func NewVerificationHandler(
	verifyPhone *usecases.VerifyPhone,
	requestEmail *usecases.RequestEmailVerification,
	verifyEmailToken *usecases.VerifyEmailToken,
) *VerificationHandler {
	return &VerificationHandler{
		verifyPhoneUC:      verifyPhone,
		requestEmailUC:     requestEmail,
		verifyEmailTokenUC: verifyEmailToken,
	}
}

type VerifyPhoneRequest struct {
	OTP string `json:"otp" binding:"required"`
}

func (h *VerificationHandler) VerifyPhone(c *gin.Context) {
	var req VerifyPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	userIDStr, _ := c.Get(authMiddleware.UserIDKey)
	userID, _ := uuid.Parse(userIDStr.(string))

	err := h.verifyPhoneUC.Execute(c.Request.Context(), usecases.VerifyPhoneRequest{
		UserID: userID,
		OTP:    req.OTP,
	})

	if err != nil {
		if errors.Is(err, domain.ErrOTPInvalid) || errors.Is(err, domain.ErrOTPNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_or_expired_otp"})
			return
		}
		if errors.Is(err, domain.ErrOTPMaxAttemptsExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "max_otp_attempts_exceeded"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_verify_phone"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "phone_verified_successfully"})
}

func (h *VerificationHandler) RequestEmailVerification(c *gin.Context) {
	userIDStr, _ := c.Get(authMiddleware.UserIDKey)
	userID, _ := uuid.Parse(userIDStr.(string))

	err := h.requestEmailUC.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_send_verification_email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification_email_sent"})
}

type VerifyEmailTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *VerificationHandler) VerifyEmailToken(c *gin.Context) {
	var req VerifyEmailTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err := h.verifyEmailTokenUC.Execute(c.Request.Context(), usecases.VerifyEmailTokenRequest{
		Token: req.Token,
	})

	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) || errors.Is(err, domain.ErrTokenExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_or_expired_token"})
			return
		}
		if errors.Is(err, domain.ErrTokenAlreadyUsed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token_already_used"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_verify_email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email_verified_successfully"})
}