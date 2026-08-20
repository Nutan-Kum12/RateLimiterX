package unit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/limiter"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/model"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client, func() {
		client.Close()
		mr.Close()
	}
}

func TestFixedWindow_AllowWithinLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	policy := &model.RateLimitPolicy{
		Algorithm: "fixed_window",
		Limit:     5,
		Window:    time.Minute,
	}

	l := limiter.NewFixedWindowLimiter(client, policy)
	ctx := context.Background()

	// Should allow 5 requests
	for i := 0; i < 5; i++ {
		result, err := l.Allow(ctx, "test-user")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d: expected allowed, got denied", i+1)
		}
		expectedRemaining := 5 - (i + 1)
		if result.Remaining != expectedRemaining {
			t.Errorf("request %d: expected remaining=%d, got=%d", i+1, expectedRemaining, result.Remaining)
		}
	}
}

func TestFixedWindow_DenyOverLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	policy := &model.RateLimitPolicy{
		Algorithm: "fixed_window",
		Limit:     3,
		Window:    time.Minute,
	}

	l := limiter.NewFixedWindowLimiter(client, policy)
	ctx := context.Background()

	// Use up all 3 requests
	for i := 0; i < 3; i++ {
		result, err := l.Allow(ctx, "test-user")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d: expected allowed", i+1)
		}
	}

	// 4th request should be denied
	result, err := l.Allow(ctx, "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected denied, got allowed")
	}
	if result.Remaining != 0 {
		t.Errorf("expected remaining=0, got=%d", result.Remaining)
	}
}

func TestFixedWindow_DifferentKeys(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	policy := &model.RateLimitPolicy{
		Algorithm: "fixed_window",
		Limit:     2,
		Window:    time.Minute,
	}

	l := limiter.NewFixedWindowLimiter(client, policy)
	ctx := context.Background()

	// User A uses 2 requests
	for i := 0; i < 2; i++ {
		l.Allow(ctx, "user-a")
	}

	// User B should still have full quota
	result, err := l.Allow(ctx, "user-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Fatal("user-b should be allowed")
	}
	if result.Remaining != 1 {
		t.Errorf("expected remaining=1 for user-b, got=%d", result.Remaining)
	}
}

func TestFixedWindow_LimitField(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	policy := &model.RateLimitPolicy{
		Algorithm: "fixed_window",
		Limit:     10,
		Window:    time.Minute,
	}

	l := limiter.NewFixedWindowLimiter(client, policy)
	ctx := context.Background()

	result, err := l.Allow(ctx, "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Limit != 10 {
		t.Errorf("expected limit=10, got=%d", result.Limit)
	}
}
