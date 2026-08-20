package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/repositories"
)

type ListPromotionsInput struct {
	Status *entities.PromotionStatus
	Limit  int
	Offset int
}

type ListPromotions struct {
	promoRepo repositories.PromotionRepository
}

func NewListPromotions(promoRepo repositories.PromotionRepository) *ListPromotions {
	return &ListPromotions{promoRepo: promoRepo}
}

func (uc *ListPromotions) Execute(ctx context.Context, input ListPromotionsInput) ([]*entities.Promotion, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	return uc.promoRepo.List(ctx, input.Status, input.Limit, input.Offset)
}
