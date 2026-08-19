package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByPhone(ctx context.Context, phone string) (*entities.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateEmailVerified(ctx context.Context, id uuid.UUID) error
	UpdatePhoneVerified(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error
	GetRolesByUserID(ctx context.Context, userID uuid.UUID) ([]string, error)
}
