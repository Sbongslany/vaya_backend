package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
)

type AdminRepository interface {
	// Phase 17: Existing Dashboard Methods
	GetPlatformStats(ctx context.Context) (*entities.PlatformStats, error)
	GetFinancialSummary(ctx context.Context, startDate, endDate time.Time) (*entities.FinancialSummary, error)
	ListUsers(ctx context.Context, limit, offset int, role, status string) ([]*entities.UserSummary, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status entities.UserStatus) error
	ListAllTrips(ctx context.Context, limit, offset int, status string) ([]*entities.TripSummary, error)

	// Phase B: New Live Operations Methods
	ListActiveTrips(ctx context.Context) ([]*entities.LiveTrip, error)
	ListOnlineDrivers(ctx context.Context) ([]*entities.LiveDriver, error)
	ForceCancelTrip(ctx context.Context, tripID uuid.UUID, reason string) error
	ForceCompleteTrip(ctx context.Context, tripID uuid.UUID) error
	ListActiveSOS(ctx context.Context) ([]*entities.LiveSOS, error)
	CreateAuditLog(ctx context.Context, log *entities.AdminAuditLog) error

	// Phase C: Payout Approvals
	ListPendingPayouts(ctx context.Context) ([]*entities.PayoutSummary, error)
	GetPayoutByID(ctx context.Context, payoutID uuid.UUID) (*entities.PayoutDetails, error)
	ApprovePayout(ctx context.Context, payoutID uuid.UUID, adminID uuid.UUID) error
	RejectPayout(ctx context.Context, payoutID uuid.UUID, adminID uuid.UUID, reason string) error
}