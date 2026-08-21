package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/repositories"
)

type AdminTopupInput struct {
	UserID      uuid.UUID
	Amount      float64
	Description string
	AdminID     uuid.UUID
}

type AdminTopup struct {
	walletRepo repositories.WalletRepository
	ledgerRepo repositories.LedgerRepository
}

func NewAdminTopup(
	walletRepo repositories.WalletRepository,
	ledgerRepo repositories.LedgerRepository,
) *AdminTopup {
	return &AdminTopup{
		walletRepo: walletRepo,
		ledgerRepo: ledgerRepo,
	}
}

func (uc *AdminTopup) Execute(ctx context.Context, input AdminTopupInput) (*entities.Wallet, error) {
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	wallet, err := uc.walletRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, domain.ErrWalletNotFound
	}

	newBalance := wallet.Balance + input.Amount

	if err := uc.walletRepo.UpdateBalance(ctx, wallet.ID, newBalance); err != nil {
		return nil, err
	}

	refType := entities.RefAdminTopup
	entry := &entities.LedgerEntry{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		EntryType:     entities.LedgerEntryCredit,
		Amount:        input.Amount,
		BalanceAfter:  newBalance,
		ReferenceType: &refType,
		ReferenceID:   &input.AdminID,
		Description:   input.Description,
		CreatedBy:     &input.AdminID,
		CreatedAt:     time.Now(),
	}

	if err := uc.ledgerRepo.Create(ctx, entry); err != nil {
		return nil, err
	}

	wallet.Balance = newBalance
	return wallet, nil
}
