package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/admin/domain/repositories"
)

// --- Get Platform Overview ---

type GetPlatformOverview struct {
	adminRepo repositories.AdminRepository
}

func NewGetPlatformOverview(adminRepo repositories.AdminRepository) *GetPlatformOverview {
	return &GetPlatformOverview{adminRepo: adminRepo}
}

func (uc *GetPlatformOverview) Execute(ctx context.Context) (*entities.PlatformStats, error) {
	return uc.adminRepo.GetPlatformStats(ctx)
}

// --- Get Financial Summary ---

type GetFinancialSummary struct {
	adminRepo repositories.AdminRepository
}

func NewGetFinancialSummary(adminRepo repositories.AdminRepository) *GetFinancialSummary {
	return &GetFinancialSummary{adminRepo: adminRepo}
}

func (uc *GetFinancialSummary) Execute(ctx context.Context) (*entities.FinancialSummary, error) {
	return uc.adminRepo.GetFinancialSummary(ctx)
}

// --- List Users ---

type ListUsersInput struct {
	Role   *entities.UserRole
	Status *entities.UserStatus
	Limit  int
	Offset int
}

type ListUsers struct {
	adminRepo repositories.AdminRepository
}

func NewListUsers(adminRepo repositories.AdminRepository) *ListUsers {
	return &ListUsers{adminRepo: adminRepo}
}

func (uc *ListUsers) Execute(ctx context.Context, input ListUsersInput) ([]*entities.UserSummary, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	return uc.adminRepo.ListUsers(ctx, input.Role, input.Status, input.Limit, input.Offset)
}

// --- Update User Status (Suspend/Ban) ---

type UpdateUserStatusInput struct {
	UserID uuid.UUID
	Status entities.UserStatus
}

type UpdateUserStatus struct {
	adminRepo repositories.AdminRepository
}

func NewUpdateUserStatus(adminRepo repositories.AdminRepository) *UpdateUserStatus {
	return &UpdateUserStatus{adminRepo: adminRepo}
}

func (uc *UpdateUserStatus) Execute(ctx context.Context, input UpdateUserStatusInput) error {
	return uc.adminRepo.UpdateUserStatus(ctx, input.UserID, input.Status)
}

// --- List All Trips ---

type ListAllTripsInput struct {
	Status *string
	Limit  int
	Offset int
}

type ListAllTrips struct {
	adminRepo repositories.AdminRepository
}

func NewListAllTrips(adminRepo repositories.AdminRepository) *ListAllTrips {
	return &ListAllTrips{adminRepo: adminRepo}
}

func (uc *ListAllTrips) Execute(ctx context.Context, input ListAllTripsInput) ([]*entities.TripSummary, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	return uc.adminRepo.ListTrips(ctx, input.Status, input.Limit, input.Offset)
}
