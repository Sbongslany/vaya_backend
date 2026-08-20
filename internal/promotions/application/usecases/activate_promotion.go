package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/repositories"
)

type ActivatePromotion struct {
	promoRepo repositories.PromotionRepository
}

func NewActivatePromotion(promoRepo repositories.PromotionRepository) *ActivatePromotion {
	return &ActivatePromotion{promoRepo: promoRepo}
}

func (uc *ActivatePromotion) Execute(ctx context.Context, id uuid.UUID) (*entities.Promotion, error) {
	promo, err := uc.promoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, domain.ErrPromotionNotFound
	}

	if promo.Status == entities.PromotionStatusActive {
		return promo, nil
	}

	if promo.Status == entities.PromotionStatusExpired {
		return nil, domain.ErrPromotionExpired
	}

	// Cannot activate if already past the valid window
	if time.Now().After(promo.ValidUntil) {
		return nil, domain.ErrPromotionExpired
	}

	if err := uc.promoRepo.UpdateStatus(ctx, id, entities.PromotionStatusActive); err != nil {
		return nil, err
	}

	promo.Status = entities.PromotionStatusActive
	return promo, nil
}
