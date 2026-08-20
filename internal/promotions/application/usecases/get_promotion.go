package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/repositories"
)

type GetPromotion struct {
	promoRepo repositories.PromotionRepository
}

func NewGetPromotion(promoRepo repositories.PromotionRepository) *GetPromotion {
	return &GetPromotion{promoRepo: promoRepo}
}

func (uc *GetPromotion) Execute(ctx context.Context, id uuid.UUID) (*entities.Promotion, error) {
	promo, err := uc.promoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, domain.ErrPromotionNotFound
	}
	return promo, nil
}
