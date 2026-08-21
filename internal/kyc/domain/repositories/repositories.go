package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/kyc/domain/entities"
)

type KYCRepository interface {
	ListPendingDrivers(ctx context.Context, limit, offset int) ([]*entities.DriverKYCSummary, error)
	GetDocumentsByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.DriverDocument, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*entities.DriverDocument, error)
	UpdateDocumentStatus(ctx context.Context, id uuid.UUID, status entities.DocumentStatus, reason *string) error
	UpdateUserOnboardingStatus(ctx context.Context, userID uuid.UUID, status entities.OnboardingStatus) error
	CountDocumentStatuses(ctx context.Context, userID uuid.UUID) (total, pending, approved, rejected int, err error)
}
