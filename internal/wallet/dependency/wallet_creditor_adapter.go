package dependency

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
)

type WalletCreditorAdapter struct {
	adminTopupUC *usecases.AdminTopup
}

func NewWalletCreditorAdapter(adminTopupUC *usecases.AdminTopup) *WalletCreditorAdapter {
	return &WalletCreditorAdapter{adminTopupUC: adminTopupUC}
}

func (a *WalletCreditorAdapter) CreditUserWallet(ctx context.Context, userID uuid.UUID, amount float64, description string, processedBy *uuid.UUID) error {
	adminID := uuid.Nil
	if processedBy != nil {
		adminID = *processedBy
	}

	_, err := a.adminTopupUC.Execute(ctx, usecases.AdminTopupInput{
		UserID:      userID,
		Amount:      amount,
		Description: description,
		AdminID:     adminID,
	})
	return err
}
