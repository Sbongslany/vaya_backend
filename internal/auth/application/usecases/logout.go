package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type LogoutUser struct {
	sessionRepo repositories.SessionRepository
}

func NewLogoutUser(sessionRepo repositories.SessionRepository) *LogoutUser {
	return &LogoutUser{sessionRepo: sessionRepo}
}

func (uc *LogoutUser) Execute(ctx context.Context, sessionID uuid.UUID) error {
	return uc.sessionRepo.Revoke(ctx, sessionID, time.Now())
}

type LogoutAllUsers struct {
	sessionRepo repositories.SessionRepository
}

func NewLogoutAllUsers(sessionRepo repositories.SessionRepository) *LogoutAllUsers {
	return &LogoutAllUsers{sessionRepo: sessionRepo}
}

func (uc *LogoutAllUsers) Execute(ctx context.Context, userID uuid.UUID) error {
	return uc.sessionRepo.RevokeAllByUserID(ctx, userID, time.Now())
}