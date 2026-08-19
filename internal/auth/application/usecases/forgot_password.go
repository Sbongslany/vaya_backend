package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/email"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/sms"
)

type ForgotPasswordRequest struct {
	Email   *string
	Phone   *string
	Channel domain.OTPChannel // SMS or EMAIL
}

type ForgotPassword struct {
	userRepo      repositories.UserRepository
	tokenRepo     repositories.PasswordResetRepository
	tokenSvc      services.TokenService
	smsProvider   sms.SMSProvider
	emailProvider email.EmailProvider
}

func NewForgotPassword(
	userRepo repositories.UserRepository,
	tokenRepo repositories.PasswordResetRepository,
	tokenSvc services.TokenService,
	smsProvider sms.SMSProvider,
	emailProvider email.EmailProvider,
) *ForgotPassword {
	return &ForgotPassword{
		userRepo: userRepo, tokenRepo: tokenRepo, tokenSvc: tokenSvc,
		smsProvider: smsProvider, emailProvider: emailProvider,
	}
}

func (uc *ForgotPassword) Execute(ctx context.Context, req ForgotPasswordRequest) error {
	var user *entities.User
	var err error
	var identifier string

	if req.Email != nil && *req.Email != "" {
		identifier = *req.Email
		user, err = uc.userRepo.FindByEmail(ctx, identifier)
	} else if req.Phone != nil && *req.Phone != "" {
		identifier = *req.Phone
		user, err = uc.userRepo.FindByPhone(ctx, identifier)
	} else {
		return fmt.Errorf("email or phone is required")
	}

	// CRITICAL: Prevent User Enumeration.
	// If the user doesn't exist, we silently return success so attackers 
	// cannot guess which emails/phones are registered in our system.
	if err == domain.ErrUserNotFound || user == nil {
		return nil 
	}
	if err != nil {
		return err
	}

	// Invalidate previous unused tokens
	if err := uc.tokenRepo.InvalidatePrevious(ctx, user.ID); err != nil {
		return err
	}

	// Generate secure token
	plainToken, err := uc.tokenSvc.GenerateSecureToken()
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	now := time.Now()
	entity := &entities.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(1 * time.Hour), // 1 hour for password resets
		CreatedAt: now,
	}

	if err := uc.tokenRepo.Create(ctx, entity); err != nil {
		return err
	}

	// Send notification
	message := fmt.Sprintf("Your password reset token is: %s. It expires in 1 hour.", plainToken)
	subject := "Password Reset Request"

	if req.Channel == domain.OTPChannelSMS {
		return uc.smsProvider.SendSMS(ctx, identifier, message)
	}
	
	return uc.emailProvider.SendEmail(ctx, identifier, subject, message)
}