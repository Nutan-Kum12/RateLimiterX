package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX/internal/model"
)

// TokenBucketLimiter implements the Token Bucket algorithm.
//
// How it works:
//   - A bucket holds tokens up to a maximum capacity (burst size).
//   - Tokens are added at a steady rate (limit / window_seconds per second).
//   - Each request consumes one token.
//   - If the bucket is empty, the request is denied.
//   - Allows short bursts above the steady-state rate up to the burst capacity.
//
// Redis implementation:
//   - Key: "ratelimit:tb:{key}" — a Redis Hash with fields: tokens, last_refill
//   - Lua script atomically calculates refill, consumes a token, and updates state.
type TokenBucketLimiter struct {
	client     *redis.Client
	limit      int           // Requests per window (used for refill rate)
	window     time.Duration // Time window for the rate
	burst      int           // Maximum bucket capacity
	refillRate float64       // Tokens per second
}

// NewTokenBucketLimiter creates a new Token Bucket limiter.
func NewTokenBucketLimiter(client *redis.Client, policy *model.RateLimitPolicy) Limiter {
	burst := policy.Burst
	if burst <= 0 {
		burst = policy.Limit // Default burst = limit
	}

	refillRate := float64(policy.Limit) / policy.Window.Seconds()

	return &TokenBucketLimiter{
		client:     client,
		limit:      policy.Limit,
		window:     policy.Window,
		burst:      burst,
		refillRate: refillRate,
	}
}

// tokenBucketScript atomically refills tokens, attempts to consume one,
// and returns the result.
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local burst = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window_seconds = tonumber(ARGV[4])

-- Get current bucket state
local tokens = tonumber(redis.call("HGET", key, "tokens"))
local last_refill = tonumber(redis.call("HGET", key, "last_refill"))

-- Initialize bucket on first request
if tokens == nil then
    tokens = burst
    last_refill = now
end

-- Calculate tokens to add since last refill
local elapsed = now - last_refill
local new_tokens = elapsed * refill_rate
tokens = math.min(burst, tokens + new_tokens)

-- Try to consume one token
local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

-- Update bucket state
redis.call("HSET", key, "tokens", tokens)
redis.call("HSET", key, "last_refill", now)
redis.call("EXPIRE", key, window_seconds * 2)

-- Calculate time until next token is available
local wait_time = 0
if allowed == 0 then
    wait_time = math.ceil((1 - tokens) / refill_rate)
end

return {allowed, math.floor(tokens), wait_time}
`)

// Allow checks if a request is allowed under the token bucket algorithm.
func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	now := time.Now()
	nowSeconds := float64(now.UnixMilli()) / 1000.0

	redisKey := fmt.Sprintf("ratelimit:tb:%s", key)
	windowSeconds := int(l.window.Seconds())

	vals, err := tokenBucketScript.Run(ctx, l.client,
		[]string{redisKey},
		l.burst, l.refillRate, nowSeconds, windowSeconds,
	).Int64Slice()
	if err != nil {
		return nil, fmt.Errorf("token bucket: redis script failed: %w", err)
	}

	if len(vals) != 3 {
		return nil, fmt.Errorf("token bucket: unexpected script response")
	}
	allowed := vals[0] == 1
	remaining := int(vals[1])
	waitTime := int(vals[2])

	resetAt := now.Add(time.Duration(waitTime) * time.Second)
	if allowed {
		resetAt = now
	}

	return &Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   resetAt,
		Limit:     l.burst,
	}, nil
}
