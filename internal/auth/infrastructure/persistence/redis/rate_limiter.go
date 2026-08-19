package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *goredis.Client
}

func NewRateLimiter(client *goredis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	
	current, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	// Set expiration on the first request in the window
	if current == 1 {
		r.client.Expire(ctx, redisKey, window)
	}

	return current <= limit, nil
}