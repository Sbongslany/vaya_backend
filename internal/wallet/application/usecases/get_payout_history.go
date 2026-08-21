package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/repositories"
)

type GetPayoutHistory struct {
	payoutRepo repositories.PayoutRepository
}

func NewGetPayoutHistory(payoutRepo repositories.PayoutRepository) *GetPayoutHistory {
	return &GetPayoutHistory{payoutRepo: payoutRepo}
}

func (uc *GetPayoutHistory) Execute(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Payout, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return uc.payoutRepo.FindByUserID(ctx, userID, limit, offset)
}
