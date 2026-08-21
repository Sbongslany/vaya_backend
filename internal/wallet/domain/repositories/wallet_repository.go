package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *entities.Wallet) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.Wallet, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Wallet, error)
	GetPlatformWallet(ctx context.Context) (*entities.Wallet, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance float64) error
}

type LedgerRepository interface {
	Create(ctx context.Context, entry *entities.LedgerEntry) error
	FindByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*entities.LedgerEntry, error)
	FindByReference(ctx context.Context, refType entities.LedgerReferenceType, refID uuid.UUID) ([]*entities.LedgerEntry, error)
}
