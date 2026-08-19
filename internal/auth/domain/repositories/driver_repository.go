package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type DriverRepository interface {
	CreateProfile(ctx context.Context, profile *entities.DriverProfile) error
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*entities.DriverProfile, error)
	GetProfileByID(ctx context.Context, id uuid.UUID) (*entities.DriverProfile, error)
	UpdateProfile(ctx context.Context, profile *entities.DriverProfile) error // NEW
	UpdateOnboardingStep(ctx context.Context, profileID uuid.UUID, step string) error
	UpdateStatus(ctx context.Context, profileID uuid.UUID, status string) error

	CreateVehicle(ctx context.Context, vehicle *entities.Vehicle) error
	GetVehiclesByProfileID(ctx context.Context, profileID uuid.UUID) ([]*entities.Vehicle, error)
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *entities.DriverDocument) error
	GetByProfileID(ctx context.Context, profileID uuid.UUID) ([]*entities.DriverDocument, error)
	UpdateStatus(ctx context.Context, docID uuid.UUID, status string, adminNotes *string) error
}

type VerificationRepository interface {
	Create(ctx context.Context, verification *entities.IdentityVerification) error
	GetLatestByProfileID(ctx context.Context, profileID uuid.UUID) (*entities.IdentityVerification, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, webhookData []byte) error
}
type DocumentRequirementRepository interface {
	GetAll(ctx context.Context) ([]*entities.DocumentRequirement, error)
	UpdateMandatoryStatus(ctx context.Context, docType string, isMandatory bool) error
}
