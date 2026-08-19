package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type AdminMFAConfirm struct {
	mfaRepo repositories.MFARepository
	mfaSvc  services.MFAService
}

func NewAdminMFAConfirm(mfaRepo repositories.MFARepository, mfaSvc services.MFAService) *AdminMFAConfirm {
	return &AdminMFAConfirm{mfaRepo: mfaRepo, mfaSvc: mfaSvc}
}

func (uc *AdminMFAConfirm) Execute(ctx context.Context, userID uuid.UUID, code string) error {
	mfaSecret, err := uc.mfaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return domain.ErrMFANotEnabled // Must call setup first
	}

	if mfaSecret.IsEnabled {
		return domain.ErrMFAAlreadyEnabled
	}

	plainSecret, err := uc.mfaSvc.DecryptSecret(mfaSecret.SecretEncrypted)
	if err != nil {
		return domain.ErrInternalServer
	}

	if !uc.mfaSvc.ValidateTOTP(plainSecret, code) {
		return domain.ErrMFAInvalidCode
	}

	return uc.mfaRepo.Enable(ctx, userID, time.Now())
}
