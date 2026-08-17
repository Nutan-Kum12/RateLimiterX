package limiter

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/logger"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/model"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/repository"
)

// Manager orchestrates rate limiting by mapping user tiers to their
// configured limiter instances. It is the single entry point for
// the middleware to check rate limits.
type Manager struct {
	limiters   map[string]Limiter                // tier → Limiter instance
	policies   map[string]*model.RateLimitPolicy // tier → policy (for metrics/info)
	policyRepo repository.PolicyRepository
}

// NewManager creates a Manager by building a Limiter for each configured tier.
// It uses the Factory to instantiate the correct algorithm per tier.
func NewManager(client *redis.Client, policyRepo repository.PolicyRepository) (*Manager, error) {
	policies := policyRepo.GetAll()
	limiters := make(map[string]Limiter, len(policies))

	for tier, policy := range policies {
		l, err := NewLimiter(policy.Algorithm, client, policy) //call factory
		if err != nil {
			return nil, fmt.Errorf("failed to create limiter for tier %q: %w", tier, err)
		}
		limiters[tier] = l

		logger.Log.Info("rate limiter initialized",
			zap.String("tier", tier),
			zap.String("algorithm", policy.Algorithm),
			zap.Int("limit", policy.Limit),
			zap.Duration("window", policy.Window),
			zap.Int("burst", policy.Burst),
		)
	}

	return &Manager{
		limiters:   limiters,
		policies:   policies,
		policyRepo: policyRepo,
	}, nil
}

// AllowRequest checks whether a request from the given user/tier should be allowed.
// It delegates to the appropriate Limiter based on the user's tier.
func (m *Manager) AllowRequest(ctx context.Context, userID, tier string) (*Result, *model.RateLimitPolicy, error) {
	limiter, ok := m.limiters[tier]
	if !ok {
		return nil, nil, fmt.Errorf("no rate limiter configured for tier: %s", tier)
	}

	policy := m.policies[tier]

	// Build the rate-limit key: "user:<userID>"
	key := fmt.Sprintf("user:%s", userID)

	result, err := limiter.Allow(ctx, key)
	if err != nil {
		return nil, policy, fmt.Errorf("rate limit check failed: %w", err)
	}

	return result, policy, nil
}

// GetPolicy returns the policy for a specific tier.
// func (m *Manager) GetPolicy(tier string) (*model.RateLimitPolicy, error) {
// 	policy, ok := m.policies[tier]
// 	if !ok {
// 		return nil, fmt.Errorf("no policy found for tier: %s", tier)
// 	}
// 	return policy, nil
// }
