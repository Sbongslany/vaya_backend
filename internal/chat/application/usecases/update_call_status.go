package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/chat/domain"
	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/chat/domain/repositories"
)

type UpdateCallStatusInput struct {
	SessionID uuid.UUID
	Status    entities.CallStatus
	Duration  int // Only used when ending the call
}

type UpdateCallStatus struct {
	callRepo repositories.CallSessionRepository
}

func NewUpdateCallStatus(callRepo repositories.CallSessionRepository) *UpdateCallStatus {
	return &UpdateCallStatus{callRepo: callRepo}
}

func (uc *UpdateCallStatus) Execute(ctx context.Context, input UpdateCallStatusInput) error {
	session, err := uc.callRepo.GetByID(ctx, input.SessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return domain.ErrCallNotFound
	}

	// Validate state transitions
	switch input.Status {
	case entities.CallStatusRinging:
		if session.Status != entities.CallStatusInitiated {
			return domain.ErrInvalidStatus
		}
	case entities.CallStatusConnected:
		if session.Status != entities.CallStatusRinging && session.Status != entities.CallStatusInitiated {
			return domain.ErrInvalidStatus
		}
	case entities.CallStatusEnded, entities.CallStatusMissed, entities.CallStatusDeclined:
		if session.Status == entities.CallStatusEnded {
			return domain.ErrInvalidStatus
		}
	}

	if input.Status == entities.CallStatusEnded || input.Status == entities.CallStatusMissed || input.Status == entities.CallStatusDeclined {
		return uc.callRepo.EndCall(ctx, input.SessionID, input.Status, input.Duration)
	}

	return uc.callRepo.UpdateStatus(ctx, input.SessionID, input.Status)
}
