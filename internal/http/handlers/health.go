package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	postgres *pgxpool.Pool
	redis    *redis.Client
}

func NewHealthHandler(postgres *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		postgres: postgres,
		redis:    redis,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	h.check(c)
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	h.check(c)
}

func (h *HealthHandler) check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	postgresUp := h.postgres.Ping(ctx) == nil
	redisUp := h.redis.Ping(ctx).Err() == nil

	status := http.StatusOK
	overall := "ok"

	if !postgresUp || !redisUp {
		status = http.StatusServiceUnavailable
		overall = "degraded"
	}

	c.JSON(status, gin.H{
		"status": overall,
		"time":   time.Now().UTC().Format(time.RFC3339),
		"checks": gin.H{
			"postgres": upDown(postgresUp),
			"redis":    upDown(redisUp),
		},
	})
}

func upDown(up bool) string {
	if up {
		return "up"
	}

	return "down"
}