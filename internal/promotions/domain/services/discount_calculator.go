package services

import (
	"math"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
)

type DiscountCalculator struct{}

func NewDiscountCalculator() *DiscountCalculator {
	return &DiscountCalculator{}
}

func (dc *DiscountCalculator) CalculateDiscount(promo *entities.Promotion, tripFare float64) float64 {
	var discount float64

	switch promo.DiscountType {
	case entities.DiscountTypePercentage:
		discount = tripFare * (promo.DiscountValue / 100.0)
		if promo.MaxDiscountAmount != nil && discount > *promo.MaxDiscountAmount {
			discount = *promo.MaxDiscountAmount
		}
	case entities.DiscountTypeFixedAmount:
		discount = promo.DiscountValue
	}

	if discount > tripFare {
		discount = tripFare
	}

	return math.Round(discount*100) / 100
}
