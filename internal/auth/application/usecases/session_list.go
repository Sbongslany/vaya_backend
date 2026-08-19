package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type ListSessions struct {
	sessionRepo repositories.SessionRepository
}

func NewListSessions(sessionRepo repositories.SessionRepository) *ListSessions {
	return &ListSessions{sessionRepo: sessionRepo}
}

func (uc *ListSessions) Execute(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error) {
	return uc.sessionRepo.FindByUserID(ctx, userID)
}