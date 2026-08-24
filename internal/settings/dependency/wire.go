package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/settings/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/settings/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/settings/interfaces/http/handlers"
)

type SettingsContainer struct {
	Handler *handlers.AdminSettingsHandler
}

func WireSettings(pgPool *pgxpool.Pool) *SettingsContainer {
	vehicleRepo := postgres.NewVehicleTypeRepository(pgPool)
	settingsRepo := postgres.NewSettingsRepository(pgPool)

	createVehicleUC := usecases.NewCreateVehicleType(vehicleRepo)
	listVehiclesUC := usecases.NewListVehicleTypes(vehicleRepo)
	updateVehicleUC := usecases.NewUpdateVehicleType(vehicleRepo)
	deleteVehicleUC := usecases.NewDeleteVehicleType(vehicleRepo)
	getSettingsUC := usecases.NewGetPlatformSettings(settingsRepo)
	updateSettingsUC := usecases.NewUpdatePlatformSettings(settingsRepo)

	handler := handlers.NewAdminSettingsHandler(
		createVehicleUC,
		listVehiclesUC,
		updateVehicleUC,
		deleteVehicleUC,
		getSettingsUC,
		updateSettingsUC,
	)

	return &SettingsContainer{
		Handler: handler,
	}
}
