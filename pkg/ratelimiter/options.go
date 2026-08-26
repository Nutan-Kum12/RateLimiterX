package ratelimiter

import "time"

// Option is a functional option for configuring a RateLimiter.
type Option func(*options)

// options holds the internal configuration for the RateLimiter.
type options struct {
	redisAddr     string
	redisPassword string
	algorithm     string
	redisDB       int
	limit         int
	window        time.Duration
	burst         int
}

// defaultOptions returns sensible defaults for a new RateLimiter.
func defaultOptions() options {
	return options{
		redisAddr:     "localhost:6379",
		redisPassword: "",
		redisDB:       0,
		algorithm:     "fixed_window",
		limit:         100,
		window:        time.Minute,
		burst:         0,
	}
}

// WithRedis configures the Redis connection for distributed rate limiting.
func WithRedis(addr, password string, db int) Option {
	return func(o *options) {
		o.redisAddr = addr
		o.redisPassword = password
		o.redisDB = db
	}
}

// WithAlgorithm sets the rate-limiting algorithm.
// Supported values: "fixed_window", "sliding_window", "sliding_log", "token_bucket".
func WithAlgorithm(algo string) Option {
	return func(o *options) {
		o.algorithm = algo
	}
}

// WithLimit sets the maximum number of requests allowed per window.
func WithLimit(limit int) Option {
	return func(o *options) {
		o.limit = limit
	}
}

// WithWindow sets the time window for rate limiting.
func WithWindow(window time.Duration) Option {
	return func(o *options) {
		o.window = window
	}
}

// WithBurst sets the burst capacity for the Token Bucket algorithm.
// For other algorithms, this value is ignored.
func WithBurst(burst int) Option {
	return func(o *options) {
		o.burst = burst
	}
}
