package entities

import (
	"time"

	"github.com/google/uuid"
)

type PromotionStatus string

const (
	PromotionStatusDraft   PromotionStatus = "DRAFT"
	PromotionStatusActive  PromotionStatus = "ACTIVE"
	PromotionStatusPaused  PromotionStatus = "PAUSED"
	PromotionStatusExpired PromotionStatus = "EXPIRED"
)

type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "PERCENTAGE"
	DiscountTypeFixedAmount DiscountType = "FIXED_AMOUNT"
)

type Promotion struct {
	ID                uuid.UUID
	Code              string
	Name              string
	Description       string
	DiscountType      DiscountType
	DiscountValue     float64
	MaxDiscountAmount *float64
	MinTripFare       float64
	UsageLimit        *int
	UsedCount         int
	PerUserLimit      int
	ValidFrom         time.Time
	ValidUntil        time.Time
	Status            PromotionStatus
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (p *Promotion) IsActiveAt(now time.Time) bool {
	return p.Status == PromotionStatusActive &&
		!now.Before(p.ValidFrom) &&
		!now.After(p.ValidUntil)
}

func (p *Promotion) HasUsageRemaining() bool {
	if p.UsageLimit == nil {
		return true
	}
	return p.UsedCount < *p.UsageLimit
}
