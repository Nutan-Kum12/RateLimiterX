package repository

import (
	"fmt"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/configs"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/model"
)

// PolicyRepository defines the interface for accessing rate-limit policies.
type PolicyRepository interface {
	GetByTier(tier string) (*model.RateLimitPolicy, error)
	GetAll() map[string]*model.RateLimitPolicy
}

// configPolicyRepository implements PolicyRepository backed by config (not DB).
// Policies are defined in config.yaml and loaded at startup.
type configPolicyRepository struct {
	policies map[string]*model.RateLimitPolicy
}

// NewPolicyRepository creates a PolicyRepository from the application configuration.
func NewPolicyRepository(tiers map[string]configs.TierConfig) PolicyRepository {
	policies := make(map[string]*model.RateLimitPolicy, len(tiers))
	for name, tier := range tiers {
		policies[name] = &model.RateLimitPolicy{
			Tier:      name,
			Algorithm: tier.Algorithm,
			Limit:     tier.Limit,
			Window:    tier.Window,
			Burst:     tier.Burst,
		}
	}
	return &configPolicyRepository{policies: policies}
}

// GetByTier returns the rate-limit policy for a specific tier.
func (r *configPolicyRepository) GetByTier(tier string) (*model.RateLimitPolicy, error) {
	policy, ok := r.policies[tier]
	if !ok {
		return nil, fmt.Errorf("no policy found for tier: %s", tier)
	}
	copied := *policy
	return &copied, nil
}

// GetAll returns all configured rate-limit policies keyed by tier name.
func (r *configPolicyRepository) GetAll() map[string]*model.RateLimitPolicy {
	// Return a copy to prevent external mutation
	result := make(map[string]*model.RateLimitPolicy, len(r.policies))
	for k, v := range r.policies {
		copied := *v
		result[k] = &copied
	}
	return result
}
