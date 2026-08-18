package limiter

import (
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/model"
)

// NewLimiter is the factory function that creates the appropriate Limiter
// implementation based on the algorithm name specified in the policy.

func NewLimiter(algorithm string, client *redis.Client, policy *model.RateLimitPolicy) (Limiter, error) {
	switch algorithm {
	case model.AlgorithmFixedWindow:
		return NewFixedWindowLimiter(client, policy), nil
	case model.AlgorithmSlidingWindow:
		return NewSlidingWindowLimiter(client, policy), nil
	case model.AlgorithmSlidingLog:
		return NewSlidingLogLimiter(client, policy), nil
	case model.AlgorithmTokenBucket:
		return NewTokenBucketLimiter(client, policy), nil
	default:
		return nil, fmt.Errorf("unknown rate-limiting algorithm: %s", algorithm)
	}
}
