package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type ResetPasswordRequest struct {
	Token       string
	NewPassword string
}

type ResetPassword struct {
	userRepo    repositories.UserRepository
	tokenRepo   repositories.PasswordResetRepository
	sessionRepo repositories.SessionRepository
	passwordSvc services.PasswordService
}

func NewResetPassword(
	userRepo repositories.UserRepository,
	tokenRepo repositories.PasswordResetRepository,
	sessionRepo repositories.SessionRepository,
	passwordSvc services.PasswordService,
) *ResetPassword {
	return &ResetPassword{
		userRepo: userRepo, tokenRepo: tokenRepo,
		sessionRepo: sessionRepo, passwordSvc: passwordSvc,
	}
}

func (uc *ResetPassword) Execute(ctx context.Context, req ResetPasswordRequest) error {
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

	// Fetch user to ensure they still exist and to check old password
	user, err := uc.userRepo.FindByID(ctx, tokenEntity.UserID)
	if err != nil {
		return err
	}

	// Enterprise rule: Prevent resetting to the exact same password
	if err := uc.passwordSvc.ComparePassword(user.PasswordHash, req.NewPassword); err == nil {
		return domain.ErrInvalidCredentials // Reusing error to indicate "cannot use current password"
	}

	// Hash new password
	newHash, err := uc.passwordSvc.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := uc.userRepo.UpdatePassword(ctx, user.ID, newHash); err != nil {
		return err
	}

	// Mark token as used
	if err := uc.tokenRepo.MarkUsed(ctx, tokenEntity.ID, time.Now()); err != nil {
		return err
	}

	// CRITICAL SECURITY: Revoke ALL active sessions for this user.
	// This forces them to log in again on all devices with the new password.
	if err := uc.sessionRepo.RevokeAllByUserID(ctx, user.ID, time.Now()); err != nil {
		return err
	}

	return nil
}
