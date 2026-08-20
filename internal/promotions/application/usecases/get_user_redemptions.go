package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/repositories"
)

type GetUserRedemptions struct {
	redemptionRepo repositories.RedemptionRepository
}

func NewGetUserRedemptions(redemptionRepo repositories.RedemptionRepository) *GetUserRedemptions {
	return &GetUserRedemptions{redemptionRepo: redemptionRepo}
}

func (uc *GetUserRedemptions) Execute(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.PromotionRedemption, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return uc.redemptionRepo.FindByUserID(ctx, userID, limit, offset)
}
