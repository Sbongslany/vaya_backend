package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type InitiateDriverOnboarding struct {
	driverRepo repositories.DriverRepository
}

func NewInitiateDriverOnboarding(driverRepo repositories.DriverRepository) *InitiateDriverOnboarding {
	return &InitiateDriverOnboarding{driverRepo: driverRepo}
}

func (uc *InitiateDriverOnboarding) Execute(ctx context.Context, userID uuid.UUID) (*entities.DriverProfile, error) {
	// Check if profile already exists
	existingProfile, err := uc.driverRepo.GetProfileByUserID(ctx, userID)
	if err == nil && existingProfile != nil {
		return existingProfile, nil // Already initiated
	}

	now := time.Now()
	profile := &entities.DriverProfile{
		ID:             uuid.New(),
		UserID:         userID,
		OnboardingStep: domain.StepProfileSetup,
		Status:         domain.DriverStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := uc.driverRepo.CreateProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}