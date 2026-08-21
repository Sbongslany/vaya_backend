package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
)

type AdminRepository interface {
	GetPlatformStats(ctx context.Context) (*entities.PlatformStats, error)
	GetFinancialSummary(ctx context.Context) (*entities.FinancialSummary, error)
	ListUsers(ctx context.Context, role *entities.UserRole, status *entities.UserStatus, limit, offset int) ([]*entities.UserSummary, error)
	ListTrips(ctx context.Context, status *string, limit, offset int) ([]*entities.TripSummary, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status entities.UserStatus) error
}
