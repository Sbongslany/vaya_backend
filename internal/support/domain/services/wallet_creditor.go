package services

import (
	"context"

	"github.com/google/uuid"
)

type WalletCreditor interface {
	CreditUserWallet(ctx context.Context, userID uuid.UUID, amount float64, description string, processedBy *uuid.UUID) error
}