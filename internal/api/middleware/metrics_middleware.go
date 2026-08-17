package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/metrics"
)

// MetricsMiddleware creates a Gin middleware that records Prometheus metrics
// for each HTTP request: total count by method/path/status and latency histogram.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Track active connections
		metrics.ActiveConnections.Inc()
		defer metrics.ActiveConnections.Dec()

		start := time.Now()
		c.Next()

		// Record metrics after request completes
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath() // Use registered route pattern, not raw URL (avoids cardinality explosion)

		if path == "" {
			path = "unmatched"
		}

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
