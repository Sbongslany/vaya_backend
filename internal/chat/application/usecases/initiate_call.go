package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/chat/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/chat/domain/services"
)

type InitiateCallInput struct {
	TripID     uuid.UUID
	CallerID   uuid.UUID
	ReceiverID uuid.UUID
}

type InitiateCallResult struct {
	SessionID uuid.UUID
	Token     string
}

type InitiateCall struct {
	callRepo      repositories.CallSessionRepository
	voiceProvider services.VoiceProvider
}

func NewInitiateCall(
	callRepo repositories.CallSessionRepository,
	voiceProvider services.VoiceProvider,
) *InitiateCall {
	return &InitiateCall{
		callRepo:      callRepo,
		voiceProvider: voiceProvider,
	}
}

func (uc *InitiateCall) Execute(ctx context.Context, input InitiateCallInput) (*InitiateCallResult, error) {
	now := time.Now()
	session := &entities.CallSession{
		ID:         uuid.New(),
		TripID:     input.TripID,
		CallerID:   input.CallerID,
		ReceiverID: input.ReceiverID,
		Status:     entities.CallStatusInitiated,
		CreatedAt:  now,
	}

	if err := uc.callRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Generate WebRTC token for the caller
	token, err := uc.voiceProvider.GenerateToken(ctx, input.CallerID.String(), input.TripID.String(), true)
	if err != nil {
		return nil, err
	}

	return &InitiateCallResult{
		SessionID: session.ID,
		Token:     token,
	}, nil
}
