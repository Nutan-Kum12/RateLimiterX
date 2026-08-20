package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/logger"
)

// LoggingMiddleware creates a Gin middleware that logs each HTTP request
// with structured fields using Zap.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique request ID
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		// Capture request start time
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process the request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		}

		// Add user ID if present
		if userID, exists := c.Get("userID"); exists {
			if userIDStr, ok := userID.(string); ok {
				fields = append(fields, zap.String("user_id", userIDStr))
			}
		}

		// Log at appropriate level based on status code
		switch {
		case statusCode >= 500:
			logger.Log.Error("server error", fields...)
		case statusCode >= 400:
			logger.Log.Warn("client error", fields...)
		default:
			logger.Log.Info("request completed", fields...)
		}
	}
}
