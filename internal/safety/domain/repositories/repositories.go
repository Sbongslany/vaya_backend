package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/safety/domain/entities"
)

type SOSRepository interface {
	Create(ctx context.Context, alert *entities.SOSAlert) error
	GetActiveByTripID(ctx context.Context, tripID uuid.UUID) (*entities.SOSAlert, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.SOSStatus, resolvedBy *uuid.UUID) error
	ListActive(ctx context.Context) ([]*entities.SOSAlert, error)
}

type ShareTokenRepository interface {
	Create(ctx context.Context, token *entities.TripShareToken) error
	GetByToken(ctx context.Context, token string) (*entities.TripShareToken, error)
}
