package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
)

type PayoutRepository interface {
	Create(ctx context.Context, payout *entities.Payout) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Payout, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Payout, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PayoutStatus, failureReason *string) error
	UpdatePaystackFields(ctx context.Context, id uuid.UUID, transferRef string, transferID string) error
}
