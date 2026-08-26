package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Nutan-Kum12/RateLimiterX/internal/dto"
	"github.com/Nutan-Kum12/RateLimiterX/internal/limiter"
	"github.com/Nutan-Kum12/RateLimiterX/internal/logger"
	"github.com/Nutan-Kum12/RateLimiterX/internal/metrics"
)

// RateLimitMiddleware creates a Gin middleware that enforces rate limits
// based on the user's tier. It reads the tier from the Gin context (set
// by AuthMiddleware), calls the Limiter Manager, and sets rate-limit headers.
func RateLimitMiddleware(manager *limiter.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from context (set by auth middleware)
		userID, exists := c.Get("userID")
		if !exists {
			// If no user in context, fall back to IP-based limiting with "free" tier
			userID = c.ClientIP()
			c.Set("tier", "free")
		}

		tier, _ := c.Get("tier")
		tierStr, ok := tier.(string)
		if !ok || tierStr == "" {
			tierStr = "free"
		}

		// Check rate limit
		userIDStr, ok2 := userID.(string)
		if !ok2 {
			userIDStr = fmt.Sprintf("%v", userID)
		}
		result, policy, err := manager.AllowRequest(c.Request.Context(), userIDStr, tierStr)
		if err != nil {
			logger.Log.Error("rate limit check failed",
				zap.String("user_id", userIDStr),
				zap.String("tier", tierStr),
				zap.Error(err),
			)
			// Fail-open: allow the request if rate limiting is broken
			c.Next()
			return
		}

		algorithmName := ""
		if policy != nil {
			algorithmName = policy.Algorithm
		}

		// Set rate-limit headers on every response
		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			// Record metrics for denied requests
			metrics.RateLimitHitsTotal.WithLabelValues(tierStr, algorithmName).Inc()

			retryAfter := time.Until(result.ResetAt).Seconds()
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)))

			logger.Log.Warn("rate limit exceeded",
				zap.String("user_id", userIDStr),
				zap.String("tier", tierStr),
				zap.String("algorithm", algorithmName),
				zap.Int("limit", result.Limit),
			)

			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.NewErrorResponse(
				"rate limit exceeded",
				"too many requests, please try again later",
			))
			return
		}

		// Record metrics for allowed requests
		metrics.RateLimitAllowedTotal.WithLabelValues(tierStr, algorithmName).Inc()

		c.Next()
	}
}
