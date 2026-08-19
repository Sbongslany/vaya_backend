package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"time"
)

type AdminMFASetup struct {
	userRepo repositories.UserRepository
	mfaRepo  repositories.MFARepository
	mfaSvc   services.MFAService
}

func NewAdminMFASetup(
	userRepo repositories.UserRepository,
	mfaRepo repositories.MFARepository,
	mfaSvc services.MFAService,
) *AdminMFASetup {
	return &AdminMFASetup{userRepo: userRepo, mfaRepo: mfaRepo, mfaSvc: mfaSvc}
}

func (uc *AdminMFASetup) Execute(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	accountName := user.Email
	if accountName == nil && user.Phone != nil {
		accountName = user.Phone
	}
	if accountName == nil {
		return "", domain.ErrInvalidTokenFormat
	}

	key, err := uc.mfaSvc.GenerateSecret(*accountName)
	if err != nil {
		return "", err
	}

	encryptedSecret, err := uc.mfaSvc.EncryptSecret(key.Secret())
	if err != nil {
		return "", err
	}

	now := time.Now()
	secretEntity := &entities.MFASecret{
		ID:              uuid.New(),
		UserID:          userID,
		SecretEncrypted: encryptedSecret,
		Method:          "TOTP",
		IsEnabled:       false, // Not enabled until confirmed
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := uc.mfaRepo.Upsert(ctx, secretEntity); err != nil {
		return "", err
	}

	return key.String(), nil // Returns the otpauth:// URI for QR code generation
}