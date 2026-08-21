package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/chat/domain/repositories"
)

type GetChatHistory struct {
	chatRepo repositories.ChatRepository
}

func NewGetChatHistory(chatRepo repositories.ChatRepository) *GetChatHistory {
	return &GetChatHistory{chatRepo: chatRepo}
}

func (uc *GetChatHistory) Execute(ctx context.Context, tripID uuid.UUID, limit, offset int) ([]*entities.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return uc.chatRepo.FindByTripID(ctx, tripID, limit, offset)
}
