package domain

import "errors"

var (
	ErrPromotionNotFound   = errors.New("promotion_not_found")
	ErrPromotionInactive   = errors.New("promotion_inactive")
	ErrPromotionExpired    = errors.New("promotion_expired")
	ErrPromotionCodeExists = errors.New("promotion_code_exists")
	ErrUsageLimitReached   = errors.New("usage_limit_reached")
	ErrPerUserLimitReached = errors.New("per_user_limit_reached")
	ErrFareBelowMinimum    = errors.New("fare_below_minimum")
	ErrInvalidDiscount     = errors.New("invalid_discount_value")
	ErrAlreadyRedeemed     = errors.New("already_redeemed_for_trip")
	ErrUnauthorized        = errors.New("unauthorized")
)
