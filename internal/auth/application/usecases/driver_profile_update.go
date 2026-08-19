package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type UpdateDriverProfileRequest struct {
	ProfileID     uuid.UUID
	LicenseNumber *string
	LicenseExpiry *time.Time
}

type UpdateDriverProfile struct {
	driverRepo repositories.DriverRepository
}

func NewUpdateDriverProfile(driverRepo repositories.DriverRepository) *UpdateDriverProfile {
	return &UpdateDriverProfile{driverRepo: driverRepo}
}

func (uc *UpdateDriverProfile) Execute(ctx context.Context, req UpdateDriverProfileRequest) error {
	profile, err := uc.driverRepo.GetProfileByID(ctx, req.ProfileID)
	if err != nil {
		return err
	}

	profile.LicenseNumber = req.LicenseNumber
	profile.LicenseExpiry = req.LicenseExpiry

	// Move to next step if profile setup is complete
	if profile.OnboardingStep == domain.StepProfileSetup {
		profile.OnboardingStep = domain.StepVehicleDetails
	}

	return uc.driverRepo.UpdateProfile(ctx, profile)
}
