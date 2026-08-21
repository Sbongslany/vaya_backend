package entities

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	Balance          float64
	Currency         string
	IsPlatformWallet bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
