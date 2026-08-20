package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
)

type PromotionRepository interface {
	Create(ctx context.Context, promo *entities.Promotion) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Promotion, error)
	GetByCode(ctx context.Context, code string) (*entities.Promotion, error)
	Update(ctx context.Context, promo *entities.Promotion) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PromotionStatus) error
	IncrementUsedCount(ctx context.Context, id uuid.UUID) error
	TryIncrementUsedCount(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, status *entities.PromotionStatus, limit, offset int) ([]*entities.Promotion, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
