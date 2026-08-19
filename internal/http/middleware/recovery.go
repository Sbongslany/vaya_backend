package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error(
					"panic recovered",
					"request_id", c.GetString(RequestIDKey),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"panic", fmt.Sprintf("%v", recovered),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_error",
					"message": "internal server error",
				})
			}
		}()

		c.Next()
	}
}