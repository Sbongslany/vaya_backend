package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
)

type AuditLog struct {
	ID        uuid.UUID
	UserID    *uuid.UUID // Nullable for unauthenticated events like failed logins
	Action    domain.AuditAction
	IPAddress *string
	DeviceID  *string
	UserAgent *string
	Metadata  []byte // JSONB
	CreatedAt time.Time
}