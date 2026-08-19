package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/email"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type RequestEmailVerification struct {
	userRepo      repositories.UserRepository
	tokenRepo     repositories.EmailVerificationRepository
	tokenSvc      services.TokenService
	emailProvider email.EmailProvider
	cfg           *config.Config
}

func NewRequestEmailVerification(
	userRepo repositories.UserRepository,
	tokenRepo repositories.EmailVerificationRepository,
	tokenSvc services.TokenService,
	emailProvider email.EmailProvider,
	cfg *config.Config,
) *RequestEmailVerification {
	return &RequestEmailVerification{
		userRepo: userRepo, tokenRepo: tokenRepo, tokenSvc: tokenSvc, emailProvider: emailProvider, cfg: cfg,
	}
}

func (uc *RequestEmailVerification) Execute(ctx context.Context, userID uuid.UUID) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Email == nil || *user.Email == "" {
		return fmt.Errorf("user has no email")
	}

	if user.EmailVerifiedAt != nil {
		return nil // Already verified
	}

	// Invalidate previous tokens
	uc.tokenRepo.InvalidatePrevious(ctx, userID)

	// Generate token
	plainToken, err := uc.tokenSvc.GenerateSecureToken()
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	now := time.Now()
	entity := &entities.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(24 * time.Hour), // 24 hours for email links
		CreatedAt: now,
	}

	if err := uc.tokenRepo.Create(ctx, entity); err != nil {
		return err
	}

	// Send email
	subject := "Verify your email address"
	body := fmt.Sprintf("Your email verification token is: %s\n\nThis token expires in 24 hours.", plainToken)
	
	return uc.emailProvider.SendEmail(ctx, *user.Email, subject, body)
}