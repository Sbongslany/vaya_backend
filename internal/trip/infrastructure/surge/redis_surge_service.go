package surge

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type RedisSurgeService struct {
	client         *redis.Client
	demandTTL      time.Duration
	baseDemand     int64
	surgeIncrement float64
	maxMultiplier  float64
}

func NewRedisSurgeService(client *redis.Client) *RedisSurgeService {
	return &RedisSurgeService{
		client:         client,
		demandTTL:      5 * time.Minute,
		baseDemand:     5,
		surgeIncrement: 0.25,
		maxMultiplier:  3.0,
	}
}

func zoneKey(lat, lng float64) string {
	return fmt.Sprintf("%.2f:%.2f", lat, lng)
}

func (s *RedisSurgeService) RecordDemand(ctx context.Context, lat, lng float64) error {
	key := "surge:demand:" + zoneKey(lat, lng)
	pipe := s.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, s.demandTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisSurgeService) GetMultiplier(ctx context.Context, lat, lng float64) (float64, error) {
	key := "surge:demand:" + zoneKey(lat, lng)
	val, err := s.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 1.0, nil
	}
	if err != nil {
		return 1.0, nil
	}

	return s.calculateMultiplier(val), nil
}

func (s *RedisSurgeService) calculateMultiplier(demand int64) float64 {
	if demand <= s.baseDemand {
		return 1.0
	}

	excess := float64(demand - s.baseDemand)
	multiplier := 1.0 + excess*s.surgeIncrement

	if multiplier > s.maxMultiplier {
		multiplier = s.maxMultiplier
	}

	return multiplier
}

func (s *RedisSurgeService) GetHeatmap(ctx context.Context) ([]services.SurgeZone, error) {
	var cursor uint64
	var zones []services.SurgeZone

	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, "surge:demand:*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			val, err := s.client.Get(ctx, key).Int64()
			if err != nil {
				continue
			}

			multiplier := s.calculateMultiplier(val)
			if multiplier <= 1.0 {
				continue
			}

			zoneKeyStr := strings.TrimPrefix(key, "surge:demand:")
			parts := strings.Split(zoneKeyStr, ":")
			if len(parts) != 2 {
				continue
			}

			lat, _ := strconv.ParseFloat(parts[0], 64)
			lng, _ := strconv.ParseFloat(parts[1], 64)

			zones = append(zones, services.SurgeZone{
				ZoneKey:    zoneKeyStr,
				Lat:        lat,
				Lng:        lng,
				Demand:     val,
				Multiplier: multiplier,
			})
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return zones, nil
}
