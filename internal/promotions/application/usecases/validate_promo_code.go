package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/services"
)

type ValidatePromoCodeInput struct {
	Code     string
	UserID   uuid.UUID
	TripFare float64
}

type ValidatePromoCodeResult struct {
	Promotion *entities.Promotion
	Discount  float64
	FinalFare float64
}

type ValidatePromoCode struct {
	promoRepo      repositories.PromotionRepository
	redemptionRepo repositories.RedemptionRepository
	discountCalc   *services.DiscountCalculator
}

func NewValidatePromoCode(
	promoRepo repositories.PromotionRepository,
	redemptionRepo repositories.RedemptionRepository,
	discountCalc *services.DiscountCalculator,
) *ValidatePromoCode {
	return &ValidatePromoCode{
		promoRepo:      promoRepo,
		redemptionRepo: redemptionRepo,
		discountCalc:   discountCalc,
	}
}

func (uc *ValidatePromoCode) Execute(ctx context.Context, input ValidatePromoCodeInput) (*ValidatePromoCodeResult, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))

	promo, err := uc.promoRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, domain.ErrPromotionNotFound
	}

	now := time.Now()

	if !promo.IsActiveAt(now) {
		if now.After(promo.ValidUntil) {
			return nil, domain.ErrPromotionExpired
		}
		return nil, domain.ErrPromotionInactive
	}

	if !promo.HasUsageRemaining() {
		return nil, domain.ErrUsageLimitReached
	}

	// Check per-user limit
	userRedemptions, err := uc.redemptionRepo.CountByUserAndPromotion(ctx, input.UserID, promo.ID)
	if err != nil {
		return nil, err
	}
	if userRedemptions >= promo.PerUserLimit {
		return nil, domain.ErrPerUserLimitReached
	}

	// Check minimum fare
	if input.TripFare < promo.MinTripFare {
		return nil, domain.ErrFareBelowMinimum
	}

	discount := uc.discountCalc.CalculateDiscount(promo, input.TripFare)

	return &ValidatePromoCodeResult{
		Promotion: promo,
		Discount:  discount,
		FinalFare: input.TripFare - discount,
	}, nil
}
