package ratelimiter

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestNew_WithDefaults(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rl, err := New(
		WithRedis(mr.Addr(), "", 0),
		WithAlgorithm("fixed_window"),
		WithLimit(10),
		WithWindow(time.Minute),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
}

func TestNew_InvalidAlgorithm(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	_, err := New(
		WithRedis(mr.Addr(), "", 0),
		WithAlgorithm("invalid"),
	)
	if err == nil {
		t.Fatal("expected error for invalid algorithm")
	}
}

func TestAllow_WithinLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rl, err := New(
		WithRedis(mr.Addr(), "", 0),
		WithAlgorithm("fixed_window"),
		WithLimit(5),
		WithWindow(time.Minute),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 5; i++ {
		result, err := rl.Allow(t.Context(), "test-key")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d: expected allowed", i+1)
		}
	}

	// 6th request should be denied
	result, err := rl.Allow(t.Context(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected denied after limit")
	}
}

func TestAllow_TokenBucket(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rl, err := New(
		WithRedis(mr.Addr(), "", 0),
		WithAlgorithm("token_bucket"),
		WithLimit(10),
		WithWindow(time.Minute),
		WithBurst(3),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should allow burst
	for i := 0; i < 3; i++ {
		result, err := rl.Allow(t.Context(), "test-key")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d: expected allowed (burst)", i+1)
		}
	}
}

func TestOptions_Defaults(t *testing.T) {
	opts := defaultOptions()
	if opts.redisAddr != "localhost:6379" {
		t.Errorf("expected default redis addr, got %s", opts.redisAddr)
	}
	if opts.algorithm != "fixed_window" {
		t.Errorf("expected default algorithm fixed_window, got %s", opts.algorithm)
	}
	if opts.limit != 100 {
		t.Errorf("expected default limit 100, got %d", opts.limit)
	}
	if opts.window != time.Minute {
		t.Errorf("expected default window 1m, got %v", opts.window)
	}
}
