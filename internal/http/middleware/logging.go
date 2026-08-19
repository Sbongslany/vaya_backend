package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logging(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
			"request_id", c.GetString(RequestIDKey),
		}

		if rawQuery != "" {
			attrs = append(attrs, "query", rawQuery)
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		if status >= 500 {
			log.Error("http request", attrs...)
			return
		}

		if status >= 400 {
			log.Warn("http request", attrs...)
			return
		}

		log.Info("http request", attrs...)
	}
}