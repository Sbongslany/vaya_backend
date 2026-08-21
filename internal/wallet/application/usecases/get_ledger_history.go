package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/repositories"
)

type GetLedgerHistory struct {
	walletRepo repositories.WalletRepository
	ledgerRepo repositories.LedgerRepository
}

func NewGetLedgerHistory(
	walletRepo repositories.WalletRepository,
	ledgerRepo repositories.LedgerRepository,
) *GetLedgerHistory {
	return &GetLedgerHistory{
		walletRepo: walletRepo,
		ledgerRepo: ledgerRepo,
	}
}

func (uc *GetLedgerHistory) Execute(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.LedgerEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, domain.ErrWalletNotFound
	}

	return uc.ledgerRepo.FindByWalletID(ctx, wallet.ID, limit, offset)
}
