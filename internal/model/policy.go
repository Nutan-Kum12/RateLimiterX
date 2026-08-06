package model

import "time"

// RateLimitPolicy defines the rate-limiting configuration for a user tier.
type RateLimitPolicy struct {
	Tier      string        `json:"tier"`
	Algorithm string        `json:"algorithm"` // "fixed_window", "sliding_window", "sliding_log", "token_bucket"
	Limit     int           `json:"limit"`     // Max requests allowed in the window
	Window    time.Duration `json:"window"`    // Time window duration
	Burst     int           `json:"burst"`     // Token bucket burst capacity (only for token_bucket)
}

// Algorithm name constants.
const (
	AlgorithmFixedWindow   = "fixed_window"
	AlgorithmSlidingWindow = "sliding_window"
	AlgorithmSlidingLog    = "sliding_log"
	AlgorithmTokenBucket   = "token_bucket"
)
