package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/admin/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/admin/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/admin/interfaces/http/handlers"
)

type AdminContainer struct {
	Handler *handlers.AdminHandler
}

func WireAdmin(pgPool *pgxpool.Pool) *AdminContainer {
	adminRepo := postgres.NewAdminRepository(pgPool)

	// Phase 17: Existing Dashboard Use Cases
	getOverviewUC := usecases.NewGetPlatformOverview(adminRepo)
	getFinancialUC := usecases.NewGetFinancialSummary(adminRepo)
	listUsersUC := usecases.NewListUsers(adminRepo)
	updateUserStatusUC := usecases.NewUpdateUserStatus(adminRepo)
	listTripsUC := usecases.NewListAllTrips(adminRepo)

	// Phase B: New Live Operations Use Cases
	getLiveMapUC := usecases.NewGetLiveMap(adminRepo)
	forceCancelUC := usecases.NewForceCancelTrip(adminRepo)
	forceCompleteUC := usecases.NewForceCompleteTrip(adminRepo)
	getActiveSOSUC := usecases.NewGetActiveSOS(adminRepo)

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
	)

	return &AdminContainer{
		Handler: handler,
	}
}
