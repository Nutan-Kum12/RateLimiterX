// Package ratelimiter provides a reusable, Redis-backed rate-limiting library
// for Go applications using the Gin web framework.
//
// It supports multiple rate-limiting algorithms (Fixed Window, Sliding Window,
// Sliding Log, Token Bucket) and can be integrated into any Gin application
// with just a few lines of code.
//
// Quick Start:
//
//	limiter, err := ratelimiter.New(
//	    ratelimiter.WithRedis("localhost:6379", "", 0),
//	    ratelimiter.WithAlgorithm("token_bucket"),
//	    ratelimiter.WithLimit(100),
//	    ratelimiter.WithWindow(time.Minute),
//	    ratelimiter.WithBurst(20),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	router := gin.Default()
//	router.Use(limiter.Middleware(ratelimiter.KeyByIP))
package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX/internal/limiter"
	"github.com/Nutan-Kum12/RateLimiterX/internal/model"
)

// KeyFunc extracts the rate-limit key from a Gin request context.
// Common implementations: KeyByIP, KeyByUserID, KeyByHeader.
type KeyFunc func(c *gin.Context) string

// KeyByIP returns the client's IP address as the rate-limit key.
func KeyByIP(c *gin.Context) string {
	return c.ClientIP()
}

// KeyByHeader returns a function that extracts the rate-limit key from a request header.
func KeyByHeader(header string) KeyFunc {
	return func(c *gin.Context) string {
		return c.GetHeader(header)
	}
}

// KeyByUserID extracts the user ID from the Gin context (set by auth middleware).
func KeyByUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return c.ClientIP() // Fallback to IP
	}
	if userIDStr, ok := userID.(string); ok {
		return userIDStr
	}
	return c.ClientIP()
}

// Result holds the outcome of a rate-limit check.
type Result struct {
	ResetAt   time.Time
	Remaining int
	Limit     int
	Allowed   bool
}

// RateLimiter is the main entry point for the rate-limiting library.
type RateLimiter struct {
	limiter limiter.Limiter
	opts    options
}

// New creates a new RateLimiter with the provided options.
func New(opts ...Option) (*RateLimiter, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     o.redisAddr,
		Password: o.redisPassword,
		DB:       o.redisDB,
	})

	// Verify Redis connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ratelimiter: failed to connect to Redis: %w", err)
	}

	// Build the policy from options
	policy := &model.RateLimitPolicy{
		Algorithm: o.algorithm,
		Limit:     o.limit,
		Window:    o.window,
		Burst:     o.burst,
	}

	// Create the limiter via factory
	l, err := limiter.NewLimiter(o.algorithm, client, policy)
	if err != nil {
		return nil, fmt.Errorf("ratelimiter: %w", err)
	}

	return &RateLimiter{
		limiter: l,
		opts:    o,
	}, nil
}

// Allow manually checks if a request for the given key is allowed.
func (rl *RateLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	result, err := rl.limiter.Allow(ctx, key)
	if err != nil {
		return nil, err
	}
	return &Result{
		Allowed:   result.Allowed,
		Remaining: result.Remaining,
		ResetAt:   result.ResetAt,
		Limit:     result.Limit,
	}, nil
}

// Middleware returns a Gin middleware that applies rate limiting.
// The keyFunc determines how to identify each client (by IP, user ID, etc.).
func (rl *RateLimiter) Middleware(keyFunc KeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			key = c.ClientIP()
		}

		result, err := rl.Allow(c.Request.Context(), key)
		if err != nil {
			// On error, allow the request (fail-open) but log the issue
			c.Next()
			return
		}

		// Set rate-limit headers on every response
		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			retryAfter := time.Until(result.ResetAt).Seconds()
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "rate limit exceeded",
				"error":   "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
