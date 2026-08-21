package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/repositories"
)

type HandleTransferWebhook struct {
	payoutRepo repositories.PayoutRepository
}

func NewHandleTransferWebhook(payoutRepo repositories.PayoutRepository) *HandleTransferWebhook {
	return &HandleTransferWebhook{payoutRepo: payoutRepo}
}

func (uc *HandleTransferWebhook) Execute(ctx context.Context, reference string, status string) error {
	// Find the payout by reference
	payout, err := uc.payoutRepo.GetByID(ctx, uuid.MustParse(reference))
	if err != nil {
		return err
	}
	if payout == nil {
		return fmt.Errorf("payout not found for reference: %s", reference)
	}

	switch status {
	case "success":
		return uc.payoutRepo.UpdateStatus(ctx, payout.ID, entities.PayoutStatusCompleted, nil)
	case "failed":
		failureReason := "Transfer failed"
		return uc.payoutRepo.UpdateStatus(ctx, payout.ID, entities.PayoutStatusFailed, &failureReason)
	case "reversed":
		failureReason := "Transfer reversed"
		return uc.payoutRepo.UpdateStatus(ctx, payout.ID, entities.PayoutStatusFailed, &failureReason)
	}

	return nil
}
