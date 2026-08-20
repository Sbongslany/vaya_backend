package services

import (
	"context"

	"github.com/google/uuid"
)

// PromotionRedeemer abstracts promo code redemption so the trip module
// doesn't depend directly on the promotions module.
type PromotionRedeemer interface {
	RedeemForTrip(ctx context.Context, code string, userID uuid.UUID, tripID uuid.UUID, tripFare float64) (discount float64, promotionID uuid.UUID, err error)
}
