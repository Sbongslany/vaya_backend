package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/repositories"
)

type GetWallet struct {
	walletRepo repositories.WalletRepository
}

func NewGetWallet(walletRepo repositories.WalletRepository) *GetWallet {
	return &GetWallet{walletRepo: walletRepo}
}

func (uc *GetWallet) Execute(ctx context.Context, userID uuid.UUID) (*entities.Wallet, error) {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, domain.ErrWalletNotFound
	}
	return wallet, nil
}
