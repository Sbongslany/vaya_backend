package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/repositories"
)

type UpdatePromotionInput struct {
	ID                uuid.UUID
	Code              string
	Name              string
	Description       string
	DiscountType      entities.DiscountType
	DiscountValue     float64
	MaxDiscountAmount *float64
	MinTripFare       float64
	UsageLimit        *int
	PerUserLimit      int
	ValidFrom         time.Time
	ValidUntil        time.Time
}

type UpdatePromotion struct {
	promoRepo repositories.PromotionRepository
}

func NewUpdatePromotion(promoRepo repositories.PromotionRepository) *UpdatePromotion {
	return &UpdatePromotion{promoRepo: promoRepo}
}

func (uc *UpdatePromotion) Execute(ctx context.Context, input UpdatePromotionInput) (*entities.Promotion, error) {
	promo, err := uc.promoRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, domain.ErrPromotionNotFound
	}

	// Only DRAFT or PAUSED promotions can be edited
	if promo.Status == entities.PromotionStatusActive {
		return nil, domain.ErrPromotionInactive
	}

	if input.DiscountValue <= 0 {
		return nil, domain.ErrInvalidDiscount
	}

	if input.DiscountType == entities.DiscountTypePercentage && input.DiscountValue > 100 {
		return nil, domain.ErrInvalidDiscount
	}

	if !input.ValidUntil.After(input.ValidFrom) {
		return nil, domain.ErrInvalidDiscount
	}

	// If code changed, check uniqueness
	newCode := strings.ToUpper(strings.TrimSpace(input.Code))
	if newCode != promo.Code {
		existing, err := uc.promoRepo.GetByCode(ctx, newCode)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, domain.ErrPromotionCodeExists
		}
	}

	promo.Code = newCode
	promo.Name = input.Name
	promo.Description = input.Description
	promo.DiscountType = input.DiscountType
	promo.DiscountValue = input.DiscountValue
	promo.MaxDiscountAmount = input.MaxDiscountAmount
	promo.MinTripFare = input.MinTripFare
	promo.UsageLimit = input.UsageLimit
	promo.PerUserLimit = input.PerUserLimit
	promo.ValidFrom = input.ValidFrom
	promo.ValidUntil = input.ValidUntil

	if err := uc.promoRepo.Update(ctx, promo); err != nil {
		return nil, err
	}

	return promo, nil
}
