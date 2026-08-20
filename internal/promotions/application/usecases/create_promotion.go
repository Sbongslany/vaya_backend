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

type CreatePromotionInput struct {
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
	CreatedBy         uuid.UUID
}

type CreatePromotion struct {
	promoRepo repositories.PromotionRepository
}

func NewCreatePromotion(promoRepo repositories.PromotionRepository) *CreatePromotion {
	return &CreatePromotion{promoRepo: promoRepo}
}

func (uc *CreatePromotion) Execute(ctx context.Context, input CreatePromotionInput) (*entities.Promotion, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if code == "" {
		return nil, domain.ErrPromotionCodeExists
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

	existing, err := uc.promoRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrPromotionCodeExists
	}

	now := time.Now()
	promo := &entities.Promotion{
		ID:                uuid.New(),
		Code:              code,
		Name:              input.Name,
		Description:       input.Description,
		DiscountType:      input.DiscountType,
		DiscountValue:     input.DiscountValue,
		MaxDiscountAmount: input.MaxDiscountAmount,
		MinTripFare:       input.MinTripFare,
		UsageLimit:        input.UsageLimit,
		UsedCount:         0,
		PerUserLimit:      input.PerUserLimit,
		ValidFrom:         input.ValidFrom,
		ValidUntil:        input.ValidUntil,
		Status:            entities.PromotionStatusDraft,
		CreatedBy:         &input.CreatedBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := uc.promoRepo.Create(ctx, promo); err != nil {
		return nil, err
	}

	return promo, nil
}
