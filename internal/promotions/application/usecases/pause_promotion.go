package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/repositories"
)

type PausePromotion struct {
	promoRepo repositories.PromotionRepository
}

func NewPausePromotion(promoRepo repositories.PromotionRepository) *PausePromotion {
	return &PausePromotion{promoRepo: promoRepo}
}

func (uc *PausePromotion) Execute(ctx context.Context, id uuid.UUID) (*entities.Promotion, error) {
	promo, err := uc.promoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, domain.ErrPromotionNotFound
	}

	if promo.Status != entities.PromotionStatusActive {
		return nil, domain.ErrPromotionInactive
	}

	if err := uc.promoRepo.UpdateStatus(ctx, id, entities.PromotionStatusPaused); err != nil {
		return nil, err
	}

	promo.Status = entities.PromotionStatusPaused
	return promo, nil
}
