package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX/internal/model"
)

// FixedWindowLimiter implements the Fixed Window Counter algorithm.
//
// How it works:
//   - Time is divided into fixed-size windows (e.g., 1-minute blocks).
//   - Each window has a counter that increments with each request.
//   - If the counter exceeds the limit, the request is denied.
//   - The counter resets when the window expires.
//
// Redis implementation:
//   - Key: "ratelimit:fw:{key}:{window_id}" where window_id = timestamp / window_seconds
//   - Uses INCR + EXPIRE atomically via a Lua script.
type FixedWindowLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

// NewFixedWindowLimiter creates a new Fixed Window limiter.
func NewFixedWindowLimiter(client *redis.Client, policy *model.RateLimitPolicy) Limiter {
	return &FixedWindowLimiter{
		client: client,
		limit:  policy.Limit,
		window: policy.Window,
	}
}

// fixedWindowScript is a Lua script that atomically increments the counter
// and sets the expiry only on the first request in a window.
var fixedWindowScript = redis.NewScript(`
local key = KEYS[1]
local window = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end

local ttl = redis.call("TTL", key)
return {current, ttl}
`)

// Allow checks if a request is allowed under the fixed window algorithm.
func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	windowSeconds := int(l.window.Seconds())
	if windowSeconds == 0 {
		windowSeconds = 1
	}

	// Calculate the current window ID
	now := time.Now()
	windowID := now.Unix() / int64(windowSeconds)
	redisKey := fmt.Sprintf("ratelimit:fw:%s:%d", key, windowID)

	// Execute the Lua script atomically
	vals, err := fixedWindowScript.Run(ctx, l.client, []string{redisKey}, l.limit, windowSeconds).Int64Slice()
	if err != nil {
		return nil, fmt.Errorf("fixed window: redis script failed: %w", err)
	}

	current := int(vals[0])
	ttl := int(vals[1])
	remaining := l.limit - current
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(time.Duration(ttl) * time.Second)

	return &Result{
		Allowed:   current <= l.limit,
		Remaining: remaining,
		ResetAt:   resetAt,
		Limit:     l.limit,
	}, nil
}
