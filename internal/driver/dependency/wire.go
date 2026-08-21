package dependency

import (
	"github.com/redis/go-redis/v9"

	"github.com/yourorg/ehailing/backend/internal/driver/application/usecases"
	redisRepo "github.com/yourorg/ehailing/backend/internal/driver/infrastructure/persistence/redis"
	"github.com/yourorg/ehailing/backend/internal/driver/interfaces/http/handlers"
)

type DriverContainer struct {
	Handler *handlers.DriverHandler
}

func WireDriver(redisClient *redis.Client) *DriverContainer {
	stateRepo := redisRepo.NewDriverStateRepository(redisClient)
	locRepo := redisRepo.NewDriverLocationRepository(redisClient)

	goOnlineUC := usecases.NewGoOnline(stateRepo)
	goOfflineUC := usecases.NewGoOffline(stateRepo, locRepo)
	updateLocUC := usecases.NewUpdateLocation(locRepo, stateRepo)
	getNearbyUC := usecases.NewGetNearbyDrivers(locRepo)

	handler := handlers.NewDriverHandler(
		goOnlineUC,
		goOfflineUC,
		updateLocUC,
		getNearbyUC,
	)

	return &DriverContainer{
		Handler: handler,
	}
}
