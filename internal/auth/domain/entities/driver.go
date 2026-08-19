package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
)

type DriverProfile struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	LicenseNumber  *string
	LicenseExpiry  *time.Time
	OnboardingStep domain.OnboardingStep
	Status         domain.DriverStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Vehicle struct {
	ID              uuid.UUID
	DriverProfileID uuid.UUID
	Make            string
	Model           string
	Year            int
	Color           string
	PlateNumber     string
	VehicleType     domain.VehicleType
	Status          domain.DocumentStatus // Reusing DocumentStatus for Vehicle Approval (PENDING/APPROVED/REJECTED)
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DriverDocument struct {
	ID              uuid.UUID
	DriverProfileID uuid.UUID
	VehicleID       *uuid.UUID // Nullable
	DocType         domain.DocumentType
	FileKey         string
	FileURL         string
	Status          domain.DocumentStatus
	AdminNotes      *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type IdentityVerification struct {
	ID                     uuid.UUID
	DriverProfileID        uuid.UUID
	Provider               string
	ProviderVerificationID *string
	Status                 domain.VerificationStatus
	WebhookData            []byte // JSONB
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
type DocumentRequirement struct {
	ID               uuid.UUID
	DocType          domain.DocumentType
	IsMandatory      bool
	AppliesToVehicle bool
	Description      *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}