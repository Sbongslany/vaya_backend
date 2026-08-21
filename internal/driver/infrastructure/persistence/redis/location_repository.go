package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yourorg/ehailing/backend/internal/driver/domain/entities"
)

const (
	geoKey      = "drivers:locations"
	locationTTL = 5 * time.Minute // If a driver stops pinging, remove them from the map
)

type DriverLocationRepository struct {
	client *redis.Client
}

func NewDriverLocationRepository(client *redis.Client) *DriverLocationRepository {
	return &DriverLocationRepository{client: client}
}

func (r *DriverLocationRepository) UpdateLocation(ctx context.Context, loc *entities.DriverLocation) error {
	// GEOADD key longitude latitude member
	// Note: Redis takes (lng, lat) order!
	err := r.client.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Longitude: loc.Longitude,
		Latitude:  loc.Latitude,
		Name:      loc.DriverID,
	}).Err()

	if err != nil {
		return err
	}

	// Store extra metadata (heading, speed) in a hash
	metaKey := fmt.Sprintf("driver:loc:meta:%s", loc.DriverID)
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, metaKey, map[string]interface{}{
		"heading":    loc.Heading,
		"speed":      loc.Speed,
		"updated_at": time.Now().Unix(),
	})
	pipe.Expire(ctx, metaKey, locationTTL)
	_, err = pipe.Exec(ctx)

	return err
}

func (r *DriverLocationRepository) GetLocation(ctx context.Context, driverID string) (*entities.DriverLocation, error) {
	// Get coordinates
	pos, err := r.client.GeoPos(ctx, geoKey, driverID).Result()
	if err != nil || len(pos) == 0 || pos[0] == nil {
		return nil, nil // Not found
	}

	loc := &entities.DriverLocation{
		DriverID:  driverID,
		Latitude:  pos[0].Latitude,
		Longitude: pos[0].Longitude,
		UpdatedAt: time.Now(),
	}

	// Get metadata
	metaKey := fmt.Sprintf("driver:loc:meta:%s", driverID)
	meta, err := r.client.HGetAll(ctx, metaKey).Result()
	if err == nil && len(meta) > 0 {
		// Parse heading and speed if needed (simplified for now)
		// In production, use strconv.ParseFloat
	}

	return loc, nil
}

func (r *DriverLocationRepository) FindNearbyDrivers(ctx context.Context, lat, lng, radiusKM float64) ([]string, error) {
	// GEORADIUS returns members within the radius
	// Note: Redis 6.2+ prefers GEOSEARCH, but GEORADIUS is widely compatible
	q := &redis.GeoRadiusQuery{
		Radius:    radiusKM,
		Unit:      "km",
		WithCoord: false,
		WithDist:  false,
		Count:     50, // Limit to 50 nearest drivers
		Sort:      "ASC",
	}

	res, err := r.client.GeoRadius(ctx, geoKey, lng, lat, q).Result()
	if err != nil {
		return nil, err
	}

	var drivers []string
	for _, loc := range res {
		drivers = append(drivers, loc.Name)
	}
	return drivers, nil
}

func (r *DriverLocationRepository) RemoveLocation(ctx context.Context, driverID string) error {
	pipe := r.client.Pipeline()
	pipe.ZRem(ctx, geoKey, driverID) // Remove from geo index
	pipe.Del(ctx, fmt.Sprintf("driver:loc:meta:%s", driverID))
	_, err := pipe.Exec(ctx)
	return err
}

func (r *DriverLocationRepository) SetTTL(ctx context.Context, driverID string, ttl time.Duration) error {
	metaKey := fmt.Sprintf("driver:loc:meta:%s", driverID)
	return r.client.Expire(ctx, metaKey, ttl).Err()
}
