package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/settings/domain/entities"
)

type VehicleTypeRepository interface {
	Create(ctx context.Context, v *entities.VehicleType) error
	Update(ctx context.Context, v *entities.VehicleType) error
	GetBySlug(ctx context.Context, slug string) (*entities.VehicleType, error)
	List(ctx context.Context) ([]*entities.VehicleType, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SettingsRepository interface {
	Get(ctx context.Context) (*entities.PlatformSettings, error)
	Update(ctx context.Context, s *entities.PlatformSettings) error
}
