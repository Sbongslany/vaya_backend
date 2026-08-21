package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
)

type CallSessionRepository interface {
	Create(ctx context.Context, session *entities.CallSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.CallSession, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.CallStatus) error
	EndCall(ctx context.Context, id uuid.UUID, status entities.CallStatus, durationSeconds int) error
}
