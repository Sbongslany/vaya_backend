package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entities.Payment) error
	GetByTripID(ctx context.Context, tripID uuid.UUID) (*entities.Payment, error)
	GetByReference(ctx context.Context, reference string) (*entities.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PaymentStatus) error
	UpdatePaystackFields(ctx context.Context, id uuid.UUID, reference string, authURL string) error
}
