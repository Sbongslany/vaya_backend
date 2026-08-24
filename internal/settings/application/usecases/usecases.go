package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/settings/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/settings/domain/repositories"
)

// --- Vehicle Types ---

type CreateVehicleTypeInput struct {
	Name       string
	Slug       string
	BaseFare   float64
	PerKmRate  float64
	PerMinRate float64
	IsActive   bool
}

type CreateVehicleType struct {
	repo repositories.VehicleTypeRepository
}

func NewCreateVehicleType(repo repositories.VehicleTypeRepository) *CreateVehicleType {
	return &CreateVehicleType{repo: repo}
}

func (uc *CreateVehicleType) Execute(ctx context.Context, input CreateVehicleTypeInput) (*entities.VehicleType, error) {
	v := &entities.VehicleType{
		ID:         uuid.New(),
		Name:       input.Name,
		Slug:       input.Slug,
		BaseFare:   input.BaseFare,
		PerKmRate:  input.PerKmRate,
		PerMinRate: input.PerMinRate,
		IsActive:   input.IsActive,
		CreatedAt:  time.Now(),
	}
	if err := uc.repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

type ListVehicleTypes struct {
	repo repositories.VehicleTypeRepository
}

func NewListVehicleTypes(repo repositories.VehicleTypeRepository) *ListVehicleTypes {
	return &ListVehicleTypes{repo: repo}
}

func (uc *ListVehicleTypes) Execute(ctx context.Context) ([]*entities.VehicleType, error) {
	return uc.repo.List(ctx)
}

type UpdateVehicleTypeInput struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	BaseFare   float64
	PerKmRate  float64
	PerMinRate float64
	IsActive   bool
}

type UpdateVehicleType struct {
	repo repositories.VehicleTypeRepository
}

func NewUpdateVehicleType(repo repositories.VehicleTypeRepository) *UpdateVehicleType {
	return &UpdateVehicleType{repo: repo}
}

func (uc *UpdateVehicleType) Execute(ctx context.Context, input UpdateVehicleTypeInput) error {
	v := &entities.VehicleType{
		ID:         input.ID,
		Name:       input.Name,
		Slug:       input.Slug,
		BaseFare:   input.BaseFare,
		PerKmRate:  input.PerKmRate,
		PerMinRate: input.PerMinRate,
		IsActive:   input.IsActive,
	}
	return uc.repo.Update(ctx, v)
}

type DeleteVehicleType struct {
	repo repositories.VehicleTypeRepository
}

func NewDeleteVehicleType(repo repositories.VehicleTypeRepository) *DeleteVehicleType {
	return &DeleteVehicleType{repo: repo}
}

func (uc *DeleteVehicleType) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

// --- Platform Settings ---

type GetPlatformSettings struct {
	repo repositories.SettingsRepository
}

func NewGetPlatformSettings(repo repositories.SettingsRepository) *GetPlatformSettings {
	return &GetPlatformSettings{repo: repo}
}

func (uc *GetPlatformSettings) Execute(ctx context.Context) (*entities.PlatformSettings, error) {
	return uc.repo.Get(ctx)
}

type UpdatePlatformSettingsInput struct {
	CommissionPercentage float64
	CancellationFee      float64
}

type UpdatePlatformSettings struct {
	repo repositories.SettingsRepository
}

func NewUpdatePlatformSettings(repo repositories.SettingsRepository) *UpdatePlatformSettings {
	return &UpdatePlatformSettings{repo: repo}
}

func (uc *UpdatePlatformSettings) Execute(ctx context.Context, input UpdatePlatformSettingsInput) error {
	current, err := uc.repo.Get(ctx)
	if err != nil {
		return err
	}
	current.CommissionPercentage = input.CommissionPercentage
	current.CancellationFee = input.CancellationFee
	return uc.repo.Update(ctx, current)
}
