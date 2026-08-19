package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type VerifyOTP struct {
	otpSvc services.OTPService
	cfg    *config.Config
}

func NewVerifyOTP(otpSvc services.OTPService, cfg *config.Config) *VerifyOTP {
	return &VerifyOTP{otpSvc: otpSvc, cfg: cfg}
}

func (uc *VerifyOTP) Execute(ctx context.Context, identifier string, purpose domain.OTPPurpose, plainOTP string) error {
	// Check attempts before verifying
	_, err := uc.otpSvc.IncrementAttempts(ctx, identifier, purpose, uc.cfg.OTPMaxAttempts)
	if err != nil {
		return err // Returns ErrOTPMaxAttemptsExceeded if limit reached
	}

	if err := uc.otpSvc.VerifyOTP(ctx, identifier, purpose, plainOTP); err != nil {
		return err // Returns ErrOTPInvalid or ErrOTPNotFound
	}

	// Success: Invalidate OTP and reset attempts
	uc.otpSvc.InvalidateOTP(ctx, identifier, purpose)
	uc.otpSvc.ResetAttempts(ctx, identifier, purpose)

	return nil
}
