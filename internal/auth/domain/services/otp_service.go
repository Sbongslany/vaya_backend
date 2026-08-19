package services

import (
	"context"
	"time"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
)

type OTPService interface {
	GenerateOTP(length int) (string, error)
	StoreOTP(ctx context.Context, identifier string, purpose domain.OTPPurpose, plainOTP string, ttl time.Duration) error
	VerifyOTP(ctx context.Context, identifier string, purpose domain.OTPPurpose, plainOTP string) error
	InvalidateOTP(ctx context.Context, identifier string, purpose domain.OTPPurpose) error
	IsInCooldown(ctx context.Context, identifier string, purpose domain.OTPPurpose) (bool, error)
	SetCooldown(ctx context.Context, identifier string, purpose domain.OTPPurpose, cooldown time.Duration) error
	IncrementAttempts(ctx context.Context, identifier string, purpose domain.OTPPurpose, maxAttempts int) (int, error)
	ResetAttempts(ctx context.Context, identifier string, purpose domain.OTPPurpose) error
}
