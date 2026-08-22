package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/geofence/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/geofence/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/geofence/interfaces/http/handlers"
)

type GeofenceContainer struct {
	Handler *handlers.GeofenceHandler
}

func WireGeofence(pgPool *pgxpool.Pool) *GeofenceContainer {
	geofenceRepo := postgres.NewGeofenceRepository(pgPool)
	assignmentRepo := postgres.NewZoneAssignmentRepository(pgPool)

	createGeofenceUC := usecases.NewCreateGeofence(geofenceRepo)
	listGeofencesUC := usecases.NewListGeofences(geofenceRepo)
	checkLocationUC := usecases.NewCheckLocationInGeofence(geofenceRepo)
	assignDriverUC := usecases.NewAssignDriverToZone(assignmentRepo)
	removeDriverUC := usecases.NewRemoveDriverFromZone(assignmentRepo)

	handler := handlers.NewGeofenceHandler(
		createGeofenceUC,
		listGeofencesUC,
		checkLocationUC,
		assignDriverUC,
		removeDriverUC,
	)

	return &GeofenceContainer{
		Handler: handler,
	}
}
