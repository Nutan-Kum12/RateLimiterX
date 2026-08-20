package limiter

import (
	"context"
	"time"
)

// Result holds the outcome of a rate-limit check.
type Result struct {
	ResetAt   time.Time // When the current window resets
	Remaining int       // How many requests remain in the current window
	Limit     int       // The configured maximum requests
	Allowed   bool      // Whether the request is permitted
}

// Limiter is the strategy interface that all rate-limiting algorithms implement.
// Each algorithm provides its own distributed implementation backed by Redis.
type Limiter interface {
	// Returns a Result with the decision and rate-limit metadata.
	Allow(ctx context.Context, key string) (*Result, error)
}
