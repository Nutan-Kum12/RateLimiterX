package unit

import (
	"context"
	"testing"
	"time"

	"github.com/Nutan-Kum12/RateLimiterX/internal/limiter"
	"github.com/Nutan-Kum12/RateLimiterX/internal/model"
)

func TestTokenBucket_AllowWithinBurst(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	policy := &model.RateLimitPolicy{
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Burst:     5,
	}

	l := limiter.NewTokenBucketLimiter(client, policy)
	ctx := context.Background()

	// Should allow burst number of requests immediately
	for i := 0; i < 5; i++ {
		result, err := l.Allow(ctx, "test-user")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d: expected allowed (burst)", i+1)
		}
	}
}

func TestTokenBucket_DenyWhenDrained(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	policy := &model.RateLimitPolicy{
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Burst:     3,
	}

	l := limiter.NewTokenBucketLimiter(client, policy)
	ctx := context.Background()

	// Drain all 3 tokens
	for i := 0; i < 3; i++ {
		result, err := l.Allow(ctx, "test-user")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d: expected allowed", i+1)
		}
	}

	// Next request should be denied (bucket drained)
	result, err := l.Allow(ctx, "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected denied after bucket drain")
	}
}

func TestTokenBucket_DifferentKeys(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	policy := &model.RateLimitPolicy{
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Burst:     2,
	}

	l := limiter.NewTokenBucketLimiter(client, policy)
	ctx := context.Background()

	// Drain user-a
	for i := 0; i < 2; i++ {
		l.Allow(ctx, "user-a")
	}

	// user-b should still have tokens
	result, err := l.Allow(ctx, "user-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Fatal("user-b should be allowed")
	}
}

func TestTokenBucket_DefaultBurst(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	// Burst = 0 should default to Limit
	policy := &model.RateLimitPolicy{
		Algorithm: "token_bucket",
		Limit:     5,
		Window:    time.Minute,
		Burst:     0,
	}

	l := limiter.NewTokenBucketLimiter(client, policy)
	ctx := context.Background()

	// Should allow 5 requests (burst defaults to limit)
	for i := 0; i < 5; i++ {
		result, err := l.Allow(ctx, "test-user")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d: expected allowed", i+1)
		}
	}
}
