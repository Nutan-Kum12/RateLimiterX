package limiter

import (
	"context"
	"time"
)

// Result holds the outcome of a rate-limit check.
type Result struct {
	Allowed   bool      // Whether the request is permitted
	Remaining int       // How many requests remain in the current window
	ResetAt   time.Time // When the current window resets
	Limit     int       // The configured maximum requests
}

// Limiter is the strategy interface that all rate-limiting algorithms implement.
// Each algorithm provides its own distributed implementation backed by Redis.
type Limiter interface {
	// Returns a Result with the decision and rate-limit metadata.
	Allow(ctx context.Context, key string) (*Result, error)
}
