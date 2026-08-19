package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type RevokeSession struct {
	sessionRepo repositories.SessionRepository
}

func NewRevokeSession(sessionRepo repositories.SessionRepository) *RevokeSession {
	return &RevokeSession{sessionRepo: sessionRepo}
}

func (uc *RevokeSession) Execute(ctx context.Context, userID, sessionID uuid.UUID) error {
	// Fetch session to ensure it belongs to the authenticated user
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return domain.ErrSessionNotFound
	}

	if session.UserID != userID {
		return domain.ErrForbidden
	}

	if session.IsRevoked() {
		return nil // Already revoked, silent success
	}

	return uc.sessionRepo.Revoke(ctx, sessionID, time.Now())
}