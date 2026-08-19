package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Session, error)
	FindByRefreshTokenHash(ctx context.Context, hash string) (*entities.Session, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID, lastUsedAt time.Time) error
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
	DeleteExpired(ctx context.Context) error
}