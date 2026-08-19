package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type VerifyPhoneRequest struct {
	UserID uuid.UUID
	OTP    string
}

type VerifyPhone struct {
	userRepo repositories.UserRepository
	otpSvc   services.OTPService
	cfg      *config.Config
}

func NewVerifyPhone(userRepo repositories.UserRepository, otpSvc services.OTPService, cfg *config.Config) *VerifyPhone {
	return &VerifyPhone{userRepo: userRepo, otpSvc: otpSvc, cfg: cfg}
}

func (uc *VerifyPhone) Execute(ctx context.Context, req VerifyPhoneRequest) error {
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	if user.Phone == nil || *user.Phone == "" {
		return errors.New("user has no phone number")
	}

	// Check attempts
	_, err = uc.otpSvc.IncrementAttempts(ctx, *user.Phone, domain.OTPPurposePhoneVerification, uc.cfg.OTPMaxAttempts)
	if err != nil {
		return err
	}

	// Verify OTP
	if err := uc.otpSvc.VerifyOTP(ctx, *user.Phone, domain.OTPPurposePhoneVerification, req.OTP); err != nil {
		return err
	}

	// Success: clear OTP state
	uc.otpSvc.InvalidateOTP(ctx, *user.Phone, domain.OTPPurposePhoneVerification)
	uc.otpSvc.ResetAttempts(ctx, *user.Phone, domain.OTPPurposePhoneVerification)

	// Update DB
	if err := uc.userRepo.UpdatePhoneVerified(ctx, user.ID); err != nil {
		return err
	}

	// Check if fully verified to activate account
	if user.EmailVerifiedAt != nil {
		uc.userRepo.UpdateStatus(ctx, user.ID, string(domain.StatusActive))
	}

	return nil
}
