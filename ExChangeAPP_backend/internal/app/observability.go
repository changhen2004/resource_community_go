package app

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ObservabilityMiddleware(slowRequestThreshold time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now()
		requestPath := ctx.FullPath()
		if requestPath == "" {
			requestPath = ctx.Request.URL.Path
		}

		ctx.Next()

		latency := time.Since(startedAt)
		statusCode := ctx.Writer.Status()
		errorText := strings.TrimSpace(ctx.Errors.String())

		level := "INFO"
		if errorText != "" || statusCode >= 500 {
			level = "ERROR"
		} else if slowRequestThreshold > 0 && latency >= slowRequestThreshold {
			level = "WARN"
		}

		log.Printf(
			"level=%s method=%s path=%s status=%d latency_ms=%d client_ip=%s user_agent=%q errors=%q",
			level,
			ctx.Request.Method,
			requestPath,
			statusCode,
			latency.Milliseconds(),
			ctx.ClientIP(),
			ctx.Request.UserAgent(),
			errorText,
		)
	}
}
