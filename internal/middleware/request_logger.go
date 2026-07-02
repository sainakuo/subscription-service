package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		args := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		if len(c.Errors) > 0 {
			args = append(args, "errors", c.Errors.String())
		}

		switch {
		case status >= 500:
			log.Error("http request completed", args...)
		case status >= 400:
			log.Warn("http request completed", args...)
		default:
			log.Info("http request completed", args...)
		}
	}
}
