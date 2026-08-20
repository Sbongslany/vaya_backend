package services

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

// PaymentProvider abstracts the external payment gateway.
// Swap the default implementation with a real gateway (Stripe, Paystack, etc.)
// via dependency injection when integrating live payments.
type PaymentProvider interface {
	Charge(ctx context.Context, amount float64, currency string, method entities.PaymentMethod) error
}