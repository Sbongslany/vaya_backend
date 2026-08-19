package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, token *entities.EmailVerificationToken) error
	FindByHash(ctx context.Context, hash string) (*entities.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time) error
	InvalidatePrevious(ctx context.Context, userID uuid.UUID) error
}

type PasswordResetRepository interface {
	Create(ctx context.Context, token *entities.PasswordResetToken) error
	FindByHash(ctx context.Context, hash string) (*entities.PasswordResetToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time) error
	InvalidatePrevious(ctx context.Context, userID uuid.UUID) error
}

type MFARepository interface {
	Upsert(ctx context.Context, secret *entities.MFASecret) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.MFASecret, error)
	Enable(ctx context.Context, userID uuid.UUID, confirmedAt time.Time) error
	Disable(ctx context.Context, userID uuid.UUID) error
}