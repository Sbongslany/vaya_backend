package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
)

type RedemptionRepository interface {
	Create(ctx context.Context, redemption *entities.PromotionRedemption) error
	CountByUserAndPromotion(ctx context.Context, userID, promotionID uuid.UUID) (int, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.PromotionRedemption, error)
	FindByTripID(ctx context.Context, tripID uuid.UUID) (*entities.PromotionRedemption, error)
}
