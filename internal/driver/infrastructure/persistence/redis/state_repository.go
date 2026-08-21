package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yourorg/ehailing/backend/internal/driver/domain/entities"
)

const (
	stateKeyPrefix = "driver:state:"
	stateTTL       = 24 * time.Hour // Auto-expire if app crashes without going offline
)

type DriverStateRepository struct {
	client *redis.Client
}

func NewDriverStateRepository(client *redis.Client) *DriverStateRepository {
	return &DriverStateRepository{client: client}
}

func (r *DriverStateRepository) SetStatus(ctx context.Context, driverID string, status entities.DriverStatus) error {
	key := stateKeyPrefix + driverID
	return r.client.Set(ctx, key, string(status), stateTTL).Err()
}

func (r *DriverStateRepository) GetStatus(ctx context.Context, driverID string) (entities.DriverStatus, error) {
	key := stateKeyPrefix + driverID
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return entities.DriverStatusOffline, nil
		}
		return "", err
	}
	return entities.DriverStatus(val), nil
}

func (r *DriverStateRepository) Delete(ctx context.Context, driverID string) error {
	key := stateKeyPrefix + driverID
	return r.client.Del(ctx, key).Err()
}
