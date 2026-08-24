package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/admin/domain/repositories"
)

// ==========================================
// PHASE 17: EXISTING DASHBOARD USE CASES
// ==========================================

type GetPlatformOverview struct{ repo repositories.AdminRepository }

func NewGetPlatformOverview(repo repositories.AdminRepository) *GetPlatformOverview {
	return &GetPlatformOverview{repo: repo}
}
func (uc *GetPlatformOverview) Execute(ctx context.Context) (*entities.PlatformStats, error) {
	return uc.repo.GetPlatformStats(ctx)
}

type GetFinancialSummary struct{ repo repositories.AdminRepository }

func NewGetFinancialSummary(repo repositories.AdminRepository) *GetFinancialSummary {
	return &GetFinancialSummary{repo: repo}
}
func (uc *GetFinancialSummary) Execute(ctx context.Context) (*entities.FinancialSummary, error) {
	return uc.repo.GetFinancialSummary(ctx, time.Time{}, time.Time{})
}

type ListUsersInput struct {
	Limit  int
	Offset int
	Role   string
	Status string
}

type ListUsers struct{ repo repositories.AdminRepository }

func NewListUsers(repo repositories.AdminRepository) *ListUsers { return &ListUsers{repo: repo} }
func (uc *ListUsers) Execute(ctx context.Context, input ListUsersInput) ([]*entities.UserSummary, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	return uc.repo.ListUsers(ctx, input.Limit, input.Offset, input.Role, input.Status)
}

type UpdateUserStatusInput struct {
	UserID uuid.UUID
	Status entities.UserStatus
}

type UpdateUserStatus struct{ repo repositories.AdminRepository }

func NewUpdateUserStatus(repo repositories.AdminRepository) *UpdateUserStatus {
	return &UpdateUserStatus{repo: repo}
}
func (uc *UpdateUserStatus) Execute(ctx context.Context, input UpdateUserStatusInput) error {
	return uc.repo.UpdateUserStatus(ctx, input.UserID, input.Status)
}

type ListAllTripsInput struct {
	Limit  int
	Offset int
	Status string
}

type ListAllTrips struct{ repo repositories.AdminRepository }

func NewListAllTrips(repo repositories.AdminRepository) *ListAllTrips {
	return &ListAllTrips{repo: repo}
}
func (uc *ListAllTrips) Execute(ctx context.Context, input ListAllTripsInput) ([]*entities.TripSummary, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	return uc.repo.ListAllTrips(ctx, input.Limit, input.Offset, input.Status)
}

// ==========================================
// PHASE B: NEW LIVE OPERATIONS USE CASES
// ==========================================

type GetLiveMap struct{ repo repositories.AdminRepository }

func NewGetLiveMap(repo repositories.AdminRepository) *GetLiveMap { return &GetLiveMap{repo: repo} }
func (uc *GetLiveMap) Execute(ctx context.Context) ([]*entities.LiveTrip, []*entities.LiveDriver, error) {
	trips, err := uc.repo.ListActiveTrips(ctx)
	if err != nil {
		return nil, nil, err
	}
	drivers, err := uc.repo.ListOnlineDrivers(ctx)
	if err != nil {
		return nil, nil, err
	}
	return trips, drivers, nil
}

type ForceCancelTripInput struct {
	AdminID uuid.UUID
	TripID  uuid.UUID
	Reason  string
}

type ForceCancelTrip struct{ repo repositories.AdminRepository }

func NewForceCancelTrip(repo repositories.AdminRepository) *ForceCancelTrip {
	return &ForceCancelTrip{repo: repo}
}
func (uc *ForceCancelTrip) Execute(ctx context.Context, input ForceCancelTripInput) error {
	if err := uc.repo.ForceCancelTrip(ctx, input.TripID, input.Reason); err != nil {
		return err
	}
	return uc.repo.CreateAuditLog(ctx, &entities.AdminAuditLog{
		ID: uuid.New(), AdminID: input.AdminID, Action: "FORCE_CANCEL_TRIP", ResourceType: "TRIP", ResourceID: &input.TripID, Details: input.Reason, CreatedAt: time.Now(),
	})
}

type ForceCompleteTripInput struct {
	AdminID uuid.UUID
	TripID  uuid.UUID
}

type ForceCompleteTrip struct{ repo repositories.AdminRepository }

func NewForceCompleteTrip(repo repositories.AdminRepository) *ForceCompleteTrip {
	return &ForceCompleteTrip{repo: repo}
}
func (uc *ForceCompleteTrip) Execute(ctx context.Context, input ForceCompleteTripInput) error {
	if err := uc.repo.ForceCompleteTrip(ctx, input.TripID); err != nil {
		return err
	}
	return uc.repo.CreateAuditLog(ctx, &entities.AdminAuditLog{
		ID: uuid.New(), AdminID: input.AdminID, Action: "FORCE_COMPLETE_TRIP", ResourceType: "TRIP", ResourceID: &input.TripID, Details: "Force completed", CreatedAt: time.Now(),
	})
}

type GetActiveSOS struct{ repo repositories.AdminRepository }

func NewGetActiveSOS(repo repositories.AdminRepository) *GetActiveSOS {
	return &GetActiveSOS{repo: repo}
}
func (uc *GetActiveSOS) Execute(ctx context.Context) ([]*entities.LiveSOS, error) {
	return uc.repo.ListActiveSOS(ctx)
}

// ==========================================
// PHASE C: PAYOUT APPROVAL USE CASES
// ==========================================

// PayoutProvider interface defines what the Admin module needs from the Paystack service
type PayoutProvider interface {
	InitiateTransfer(ctx context.Context, bankCode, accountNumber string, amount float64) error
}

type GetPendingPayouts struct{ repo repositories.AdminRepository }

func NewGetPendingPayouts(repo repositories.AdminRepository) *GetPendingPayouts {
	return &GetPendingPayouts{repo: repo}
}
func (uc *GetPendingPayouts) Execute(ctx context.Context) ([]*entities.PayoutSummary, error) {
	return uc.repo.ListPendingPayouts(ctx)
}

type RejectPayoutInput struct {
	AdminID  uuid.UUID
	PayoutID uuid.UUID
	Reason   string
}
type RejectPayout struct{ repo repositories.AdminRepository }

func NewRejectPayout(repo repositories.AdminRepository) *RejectPayout {
	return &RejectPayout{repo: repo}
}
func (uc *RejectPayout) Execute(ctx context.Context, input RejectPayoutInput) error {
	return uc.repo.RejectPayout(ctx, input.PayoutID, input.AdminID, input.Reason)
}

// ProcessPayoutApproval handles fetching details, calling Paystack, and updating DB
type ProcessPayoutApprovalInput struct {
	AdminID  uuid.UUID
	PayoutID uuid.UUID
}

type ProcessPayoutApproval struct {
	repo          repositories.AdminRepository
	payoutService PayoutProvider
}

func NewProcessPayoutApproval(repo repositories.AdminRepository, payoutService PayoutProvider) *ProcessPayoutApproval {
	return &ProcessPayoutApproval{repo: repo, payoutService: payoutService}
}

func (uc *ProcessPayoutApproval) Execute(ctx context.Context, input ProcessPayoutApprovalInput) error {
	// 1. Fetch payout details
	payout, err := uc.repo.GetPayoutByID(ctx, input.PayoutID)
	if err != nil {
		return err
	}
	if payout == nil {
		return fmt.Errorf("payout not found")
	}
	if payout.Status != "PENDING" {
		return fmt.Errorf("payout is not in pending status")
	}

	// 2. Call Paystack to initiate the bank transfer
	if err := uc.payoutService.InitiateTransfer(ctx, payout.BankCode, payout.AccountNumber, payout.Amount); err != nil {
		return fmt.Errorf("failed to initiate paystack transfer: %w", err)
	}

	// 3. Update database status to APPROVED
	return uc.repo.ApprovePayout(ctx, input.PayoutID, input.AdminID)
}
