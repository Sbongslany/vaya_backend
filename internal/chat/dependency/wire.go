package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/chat/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/chat/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/chat/infrastructure/voice"
	"github.com/yourorg/ehailing/backend/internal/chat/infrastructure/websocket"
	"github.com/yourorg/ehailing/backend/internal/chat/interfaces/http/handlers"
)

type ChatContainer struct {
	Handler *handlers.ChatHandler
}

func WireChat(pgPool *pgxpool.Pool) *ChatContainer {
	// Repositories
	chatRepo := postgres.NewChatRepository(pgPool)
	callRepo := postgres.NewCallSessionRepository(pgPool)

	// Services
	voiceProvider := voice.NewMockVoiceProvider()

	// Use cases
	sendMessageUC := usecases.NewSendMessage(chatRepo)
	getHistoryUC := usecases.NewGetChatHistory(chatRepo)
	initiateCallUC := usecases.NewInitiateCall(callRepo, voiceProvider)
	updateCallUC := usecases.NewUpdateCallStatus(callRepo)

	// WebSocket Hub
	chatHub := websocket.NewChatHub()

	// Handler
	handler := handlers.NewChatHandler(
		chatHub,
		sendMessageUC,
		getHistoryUC,
		initiateCallUC,
		updateCallUC,
	)

	return &ChatContainer{
		Handler: handler,
	}
}
