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

type RedeemPromotionInput struct {
	Code     string
	UserID   uuid.UUID
	TripID   uuid.UUID
	TripFare float64
}

type RedeemPromotionResult struct {
	Discount    float64
	PromotionID uuid.UUID
}

type RedeemPromotion struct {
	promoRepo      repositories.PromotionRepository
	redemptionRepo repositories.RedemptionRepository
	discountCalc   *services.DiscountCalculator
}

func NewRedeemPromotion(
	promoRepo repositories.PromotionRepository,
	redemptionRepo repositories.RedemptionRepository,
	discountCalc *services.DiscountCalculator,
) *RedeemPromotion {
	return &RedeemPromotion{
		promoRepo:      promoRepo,
		redemptionRepo: redemptionRepo,
		discountCalc:   discountCalc,
	}
}

func (uc *RedeemPromotion) Execute(ctx context.Context, input RedeemPromotionInput) (*RedeemPromotionResult, error) {
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

	userRedemptions, err := uc.redemptionRepo.CountByUserAndPromotion(ctx, input.UserID, promo.ID)
	if err != nil {
		return nil, err
	}
	if userRedemptions >= promo.PerUserLimit {
		return nil, domain.ErrPerUserLimitReached
	}

	if input.TripFare < promo.MinTripFare {
		return nil, domain.ErrFareBelowMinimum
	}

	incremented, err := uc.promoRepo.TryIncrementUsedCount(ctx, promo.ID)
	if err != nil {
		return nil, err
	}
	if !incremented {
		return nil, domain.ErrUsageLimitReached
	}

	discount := uc.discountCalc.CalculateDiscount(promo, input.TripFare)

	redemption := &entities.PromotionRedemption{
		ID:              uuid.New(),
		PromotionID:     promo.ID,
		UserID:          input.UserID,
		TripID:          &input.TripID,
		DiscountApplied: discount,
		RedeemedAt:      now,
	}

	if err := uc.redemptionRepo.Create(ctx, redemption); err != nil {
		return nil, err
	}

	return &RedeemPromotionResult{
		Discount:    discount,
		PromotionID: promo.ID,
	}, nil
}
