package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/safety/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/safety/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/safety/interfaces/http/handlers"
	tripPostgres "github.com/yourorg/ehailing/backend/internal/trip/infrastructure/persistence/postgres"
)

type SafetyContainer struct {
	Handler *handlers.SafetyHandler
}

func WireSafety(pgPool *pgxpool.Pool, baseURL string) *SafetyContainer {
	sosRepo := postgres.NewSOSRepository(pgPool)
	shareRepo := postgres.NewShareTokenRepository(pgPool)
	tripRepo := tripPostgres.NewTripRepository(pgPool)

	triggerSOSUC := usecases.NewTriggerSOS(sosRepo, tripRepo)
	generateShareUC := usecases.NewGenerateShareLink(shareRepo, tripRepo, baseURL)
	viewSharedTripUC := usecases.NewViewSharedTrip(shareRepo, tripRepo)

	handler := handlers.NewSafetyHandler(
		triggerSOSUC,
		generateShareUC,
		viewSharedTripUC,
	)

	return &SafetyContainer{
		Handler: handler,
	}
}
