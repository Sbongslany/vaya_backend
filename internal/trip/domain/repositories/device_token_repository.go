package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type DeviceTokenRepository interface {
	Save(ctx context.Context, token *entities.DeviceToken) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.DeviceToken, error)
	Delete(ctx context.Context, userID uuid.UUID, token string) error
}
