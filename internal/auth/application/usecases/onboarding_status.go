package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type GetOnboardingStatus struct {
	driverRepo repositories.DriverRepository
	reqRepo    repositories.DocumentRequirementRepository
}

func NewGetOnboardingStatus(driverRepo repositories.DriverRepository, reqRepo repositories.DocumentRequirementRepository) *GetOnboardingStatus {
	return &GetOnboardingStatus{driverRepo: driverRepo, reqRepo: reqRepo}
}

type OnboardingStatusResponse struct {
	Profile      *entities.DriverProfile         `json:"profile"`
	Vehicles     []*entities.Vehicle             `json:"vehicles"`
	Requirements []*entities.DocumentRequirement `json:"document_requirements"`
}

func (uc *GetOnboardingStatus) Execute(ctx context.Context, userID uuid.UUID) (*OnboardingStatusResponse, error) {
	profile, err := uc.driverRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	vehicles, err := uc.driverRepo.GetVehiclesByProfileID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	requirements, err := uc.reqRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return &OnboardingStatusResponse{
		Profile:      profile,
		Vehicles:     vehicles,
		Requirements: requirements,
	}, nil
}
