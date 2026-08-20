package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/api/middleware"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/auth"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/configs"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/limiter"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/logger"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/model"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/repository"
)

func init() {
	logger.Init("test")
}

// setupIntegrationTest creates a test Gin engine with auth + ratelimit middleware.
func setupIntegrationTest(t *testing.T, limit int) (*gin.Engine, *auth.JWTManager, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// Setup miniredis
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Setup JWT manager
	jwtManager := auth.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)

	// Setup policy repository with test config
	tiers := map[string]configs.TierConfig{
		"free": {
			Algorithm: "fixed_window",
			Limit:     limit,
			Window:    time.Minute,
		},
	}
	policyRepo := repository.NewPolicyRepository(tiers)

	// Setup limiter manager
	manager, err := limiter.NewManager(client, policyRepo)
	if err != nil {
		t.Fatalf("failed to create limiter manager: %v", err)
	}

	// Build Gin router
	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtManager))
	router.Use(middleware.RateLimitMiddleware(manager))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return router, jwtManager, cleanup
}

func TestMiddleware_AuthRequired(t *testing.T) {
	router, _, cleanup := setupIntegrationTest(t, 10)
	defer cleanup()

	// Request without auth header
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMiddleware_ValidAuth(t *testing.T) {
	router, jwtManager, cleanup := setupIntegrationTest(t, 10)
	defer cleanup()

	// Generate a valid token
	token, err := jwtManager.GenerateAccessToken("user-123", "test@test.com", "free")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Check rate-limit headers are present
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("missing X-RateLimit-Limit header")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("missing X-RateLimit-Remaining header")
	}
	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("missing X-RateLimit-Reset header")
	}
}

func TestMiddleware_RateLimitExceeded(t *testing.T) {
	limit := 3
	router, jwtManager, cleanup := setupIntegrationTest(t, limit)
	defer cleanup()

	token, err := jwtManager.GenerateAccessToken("user-123", "test@test.com", "free")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Send requests up to the limit
	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Next request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}

	// Verify Retry-After header
	if w.Header().Get("Retry-After") == "" {
		t.Error("missing Retry-After header on 429 response")
	}

	// Verify response body
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	if body["success"] != false {
		t.Error("expected success=false in 429 response")
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	router, _, cleanup := setupIntegrationTest(t, 10)
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	router, _, cleanup := setupIntegrationTest(t, 10)
	defer cleanup()

	// Create a JWT manager with very short TTL
	shortJWT := auth.NewJWTManager("test-access-secret", "test-refresh-secret", -time.Hour, 24*time.Hour)
	token, _ := shortJWT.GenerateAccessToken("user-123", "test@test.com", "free")

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", w.Code)
	}
}

// Ensure model package is used
var _ = model.TierFree
