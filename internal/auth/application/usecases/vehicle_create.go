package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type CreateVehicleRequest struct {
	DriverProfileID uuid.UUID
	Make            string
	Model           string
	Year            int
	Color           string
	PlateNumber     string
	VehicleType     domain.VehicleType
}

type CreateVehicle struct {
	driverRepo repositories.DriverRepository
}

func NewCreateVehicle(driverRepo repositories.DriverRepository) *CreateVehicle {
	return &CreateVehicle{driverRepo: driverRepo}
}

func (uc *CreateVehicle) Execute(ctx context.Context, req CreateVehicleRequest) (*entities.Vehicle, error) {
	now := time.Now()
	vehicle := &entities.Vehicle{
		ID:              uuid.New(),
		DriverProfileID: req.DriverProfileID,
		Make:            req.Make,
		Model:           req.Model,
		Year:            req.Year,
		Color:           req.Color,
		PlateNumber:     req.PlateNumber,
		VehicleType:     req.VehicleType,
		Status:          domain.DocStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := uc.driverRepo.CreateVehicle(ctx, vehicle); err != nil {
		return nil, err
	}

	// Move onboarding step to DOCUMENTS
	uc.driverRepo.UpdateOnboardingStep(ctx, req.DriverProfileID, string(domain.StepDocuments))

	return vehicle, nil
}