package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX/internal/dto"
)

// HealthHandler handles health and metrics endpoints.
type HealthHandler struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewHealthHandler(db *sql.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
	}
}

// Returns the health status of all dependent services.
func (h *HealthHandler) Health(c *gin.Context) {
	services := make(map[string]string)
	overallStatus := "healthy"

	// Check MySQL
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		services["mysql"] = "unhealthy: " + err.Error()
		overallStatus = "degraded"
	} else {
		services["mysql"] = "healthy"
	}

	// Check Redis
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		services["redis"] = "unhealthy: " + err.Error()
		overallStatus = "degraded"
	} else {
		services["redis"] = "healthy"
	}

	statusCode := http.StatusOK
	if overallStatus != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, dto.NewSuccessResponse(overallStatus, dto.HealthResponse{
		Status:   overallStatus,
		Services: services,
	}))
}

// Metrics handles GET /metrics
// Exposes Prometheus metrics using the default promhttp handler.
func (h *HealthHandler) Metrics() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
