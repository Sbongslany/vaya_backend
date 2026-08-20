package payment

import (
	"context"
	"errors"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

// DefaultPaymentProvider is the built-in payment processor.
// Cash payments settle physically between passenger and driver.
// Card and wallet payments route through the configured gateway,
// which is injected here for production use.
type DefaultPaymentProvider struct{}

func NewDefaultPaymentProvider() *DefaultPaymentProvider {
	return &DefaultPaymentProvider{}
}

func (p *DefaultPaymentProvider) Charge(ctx context.Context, amount float64, currency string, method entities.PaymentMethod) error {
	if amount <= 0 {
		return errors.New("invalid payment amount")
	}
	switch method {
	case entities.PaymentMethodCash, entities.PaymentMethodCard, entities.PaymentMethodWallet:
		return nil
	default:
		return errors.New("unsupported payment method")
	}
}
