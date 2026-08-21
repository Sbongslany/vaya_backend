package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/chat/domain/repositories"
)

type SendMessageInput struct {
	TripID     uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
}

type SendMessage struct {
	chatRepo repositories.ChatRepository
}

func NewSendMessage(chatRepo repositories.ChatRepository) *SendMessage {
	return &SendMessage{chatRepo: chatRepo}
}

func (uc *SendMessage) Execute(ctx context.Context, input SendMessageInput) (*entities.ChatMessage, error) {
	now := time.Now()
	message := &entities.ChatMessage{
		ID:         uuid.New(),
		TripID:     input.TripID,
		SenderID:   input.SenderID,
		ReceiverID: input.ReceiverID,
		Content:    input.Content,
		CreatedAt:  now,
	}

	if err := uc.chatRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}
