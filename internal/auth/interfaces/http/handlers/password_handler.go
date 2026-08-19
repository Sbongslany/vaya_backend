package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/ehailing/backend/internal/auth/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
)

type PasswordHandler struct {
	forgotPasswordUC *usecases.ForgotPassword
	resetPasswordUC  *usecases.ResetPassword
}

func NewPasswordHandler(forgot *usecases.ForgotPassword, reset *usecases.ResetPassword) *PasswordHandler {
	return &PasswordHandler{forgotPasswordUC: forgot, resetPasswordUC: reset}
}

type ForgotPasswordRequest struct {
	Email   *string           `json:"email"`
	Phone   *string           `json:"phone"`
	Channel domain.OTPChannel `json:"channel" binding:"required"`
}

func (h *PasswordHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email_or_phone_required"})
		return
	}

	err := h.forgotPasswordUC.Execute(c.Request.Context(), usecases.ForgotPasswordRequest{
		Email:   req.Email,
		Phone:   req.Phone,
		Channel: req.Channel,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_process_request"})
		return
	}

	// Always return success to prevent user enumeration
	c.JSON(http.StatusOK, gin.H{
		"message": "if_the_account_exists_instructions_have_been_sent",
	})
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *PasswordHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	err := h.resetPasswordUC.Execute(c.Request.Context(), usecases.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
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
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot_use_current_password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_reset_password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password_reset_successfully"})
}