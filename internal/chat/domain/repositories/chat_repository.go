package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
)

type ChatRepository interface {
	Create(ctx context.Context, message *entities.ChatMessage) error
	FindByTripID(ctx context.Context, tripID uuid.UUID, limit, offset int) ([]*entities.ChatMessage, error)
	MarkAsRead(ctx context.Context, tripID uuid.UUID, readerID uuid.UUID) error
}
