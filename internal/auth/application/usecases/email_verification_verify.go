package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type VerifyEmailTokenRequest struct {
	Token string
}

type VerifyEmailToken struct {
	userRepo  repositories.UserRepository
	tokenRepo repositories.EmailVerificationRepository
}

func NewVerifyEmailToken(userRepo repositories.UserRepository, tokenRepo repositories.EmailVerificationRepository) *VerifyEmailToken {
	return &VerifyEmailToken{userRepo: userRepo, tokenRepo: tokenRepo}
}

func (uc *VerifyEmailToken) Execute(ctx context.Context, req VerifyEmailTokenRequest) error {
	hash := sha256.Sum256([]byte(req.Token))
	tokenHash := hex.EncodeToString(hash[:])

	tokenEntity, err := uc.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrTokenNotFound
	}

	if tokenEntity.IsUsed() {
		return domain.ErrTokenAlreadyUsed
	}

	if time.Now().After(tokenEntity.ExpiresAt) {
		return domain.ErrTokenExpired
	}

	// Mark used
	if err := uc.tokenRepo.MarkUsed(ctx, tokenEntity.ID, time.Now()); err != nil {
		return err
	}

	// Update user
	if err := uc.userRepo.UpdateEmailVerified(ctx, tokenEntity.UserID); err != nil {
		return err
	}

	// Fetch user to check if we should activate
	user, err := uc.userRepo.FindByID(ctx, tokenEntity.UserID)
	if err == nil && user.PhoneVerifiedAt != nil {
		uc.userRepo.UpdateStatus(ctx, user.ID, string(domain.StatusActive))
	}

	return nil
}