package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type RateLimitMiddleware struct {
	limiter services.RateLimiter
}

func NewRateLimitMiddleware(limiter services.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Limit(endpoint string, limit int64, window time.Duration, keyFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			key = c.ClientIP() // Fallback to IP if custom key func returns empty
		}

		allowed, err := m.limiter.Allow(c.Request.Context(), endpoint+":"+key, limit, window)
		if err != nil {
			// Fail closed (block request) if Redis is down to prevent brute force
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate_limit_check_failed"})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate_limit_exceeded"})
			return
		}

		c.Next()
	}
}
