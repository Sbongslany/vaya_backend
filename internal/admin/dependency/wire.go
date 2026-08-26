// package dependency

// import (
// 	"github.com/jackc/pgx/v5/pgxpool"

// 	"github.com/yourorg/ehailing/backend/internal/admin/application/usecases"
// 	"github.com/yourorg/ehailing/backend/internal/admin/infrastructure/persistence/postgres"
// 	"github.com/yourorg/ehailing/backend/internal/admin/interfaces/http/handlers"
// 	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
// )

// type AdminContainer struct {
// 	Handler *handlers.AdminHandler
// }

// func WireAdmin(
// 	pgPool *pgxpool.Pool,
// 	payoutService usecases.PayoutProvider,
// 	passwordSvc services.PasswordService,
// ) *AdminContainer {
// 	adminRepo := postgres.NewAdminRepository(pgPool)

// 	// Phase 17: Dashboard
// 	getOverviewUC := usecases.NewGetPlatformOverview(adminRepo)
// 	getFinancialUC := usecases.NewGetFinancialSummary(adminRepo)
// 	listUsersUC := usecases.NewListUsers(adminRepo)
// 	updateUserStatusUC := usecases.NewUpdateUserStatus(adminRepo)
// 	listTripsUC := usecases.NewListAllTrips(adminRepo)

// 	// Phase B: Live Operations
// 	getLiveMapUC := usecases.NewGetLiveMap(adminRepo)
// 	forceCancelUC := usecases.NewForceCancelTrip(adminRepo)
// 	forceCompleteUC := usecases.NewForceCompleteTrip(adminRepo)
// 	getActiveSOSUC := usecases.NewGetActiveSOS(adminRepo)

// 	// Phase C: Payout Approvals
// 	getPendingPayoutsUC := usecases.NewGetPendingPayouts(adminRepo)
// 	processPayoutApprovalUC := usecases.NewProcessPayoutApproval(adminRepo, payoutService)
// 	rejectPayoutUC := usecases.NewRejectPayout(adminRepo)

// 	// Phase 3: Admin Management
// 	adminMgmtUCs := usecases.NewAdminManagementUCs(adminRepo, passwordSvc)

// 	handler := handlers.NewAdminHandler(
// 		getOverviewUC,
// 		getFinancialUC,
// 		listUsersUC,
// 		updateUserStatusUC,
// 		listTripsUC,
// 		getLiveMapUC,
// 		forceCancelUC,
// 		forceCompleteUC,
// 		getActiveSOSUC,
// 		getPendingPayoutsUC,
// 		processPayoutApprovalUC,
// 		rejectPayoutUC,
// 		adminMgmtUCs,
// 	)

// 	return &AdminContainer{
// 		Handler: handler,
// 	}
// }

package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/admin/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/admin/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/admin/interfaces/http/handlers"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type AdminContainer struct {
	Handler *handlers.AdminHandler
}

func WireAdmin(
	pgPool *pgxpool.Pool,
	payoutService usecases.PayoutProvider,
	passwordSvc services.PasswordService,
) *AdminContainer {
	adminRepo := postgres.NewAdminRepository(pgPool)

	// Phase 17: Dashboard
	getOverviewUC := usecases.NewGetPlatformOverview(adminRepo)
	getFinancialUC := usecases.NewGetFinancialSummary(adminRepo)
	listUsersUC := usecases.NewListUsers(adminRepo)
	updateUserStatusUC := usecases.NewUpdateUserStatus(adminRepo)
	listTripsUC := usecases.NewListAllTrips(adminRepo)

	// Phase B: Live Operations
	getLiveMapUC := usecases.NewGetLiveMap(adminRepo)
	forceCancelUC := usecases.NewForceCancelTrip(adminRepo)
	forceCompleteUC := usecases.NewForceCompleteTrip(adminRepo)
	getActiveSOSUC := usecases.NewGetActiveSOS(adminRepo)

	// Phase C: Payout Approvals
	getPendingPayoutsUC := usecases.NewGetPendingPayouts(adminRepo)
	processPayoutApprovalUC := usecases.NewProcessPayoutApproval(adminRepo, payoutService)
	rejectPayoutUC := usecases.NewRejectPayout(adminRepo)

	// Phase 3: Admin Management
	adminMgmtUCs := usecases.NewAdminManagementUCs(adminRepo, passwordSvc)

	handler := handlers.NewAdminHandler(
		getOverviewUC,
		getFinancialUC,
		listUsersUC,
		updateUserStatusUC,
		listTripsUC,
		getLiveMapUC,
		forceCancelUC,
		forceCompleteUC,
		getActiveSOSUC,
		getPendingPayoutsUC,
		processPayoutApprovalUC,
		rejectPayoutUC,
		adminMgmtUCs,
	)

	return &AdminContainer{
		Handler: handler,
	}
}
