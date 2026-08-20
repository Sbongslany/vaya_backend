package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type RegisterDeviceTokenInput struct {
	UserID     uuid.UUID
	Token      string
	DeviceType string
}

type RegisterDeviceToken struct {
	tokenRepo repositories.DeviceTokenRepository
}

func NewRegisterDeviceToken(tokenRepo repositories.DeviceTokenRepository) *RegisterDeviceToken {
	return &RegisterDeviceToken{tokenRepo: tokenRepo}
}

func (uc *RegisterDeviceToken) Execute(ctx context.Context, input RegisterDeviceTokenInput) error {
	token := &entities.DeviceToken{
		UserID:     input.UserID,
		Token:      input.Token,
		DeviceType: input.DeviceType,
		CreatedAt:  time.Now(),
	}
	return uc.tokenRepo.Save(ctx, token)
}
