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

	getOverviewUC := usecases.NewGetPlatformOverview(adminRepo)
	getFinancialUC := usecases.NewGetFinancialSummary(adminRepo)
	listUsersUC := usecases.NewListUsers(adminRepo)
	updateUserStatusUC := usecases.NewUpdateUserStatus(adminRepo)
	listTripsUC := usecases.NewListAllTrips(adminRepo)

	handler := handlers.NewAdminHandler(
		getOverviewUC,
		getFinancialUC,
		listUsersUC,
		updateUserStatusUC,
		listTripsUC,
	)

	return &AdminContainer{
		Handler: handler,
	}
}
