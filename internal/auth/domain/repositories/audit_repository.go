package repositories

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type AuditRepository interface {
	Log(ctx context.Context, log *entities.AuditLog) error
}