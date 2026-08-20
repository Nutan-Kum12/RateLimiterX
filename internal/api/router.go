package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/api/handler"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/api/middleware"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/auth"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/limiter"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/service"
)

// RouterDeps contains all dependencies needed to set up the router.
type RouterDeps struct {
	AuthService    service.AuthService
	UserService    service.UserService
	JWTManager     *auth.JWTManager
	LimiterManager *limiter.Manager
	DB             *sql.DB
	RedisClient    *redis.Client
}

// NewRouter creates and configures the Gin engine with all routes and middleware.
func NewRouter(mode string, deps RouterDeps) *gin.Engine {
	gin.SetMode(mode)
	router := gin.New()

	// Recovery middleware — catches panics and returns 500
	router.Use(gin.Recovery())

	// Global middleware (applied to all routes)
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.MetricsMiddleware())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(deps.AuthService)
	userHandler := handler.NewUserHandler(deps.UserService)
	healthHandler := handler.NewHealthHandler(deps.DB, deps.RedisClient)

	// Public Routes
	router.GET("/health", healthHandler.Health)
	router.GET("/metrics", healthHandler.Metrics())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}
		// Protected toutes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(deps.JWTManager))
		protected.Use(middleware.RateLimitMiddleware(deps.LimiterManager))
		{
			// User routes
			userGroup := protected.Group("/users")
			{
				userGroup.GET("/me", userHandler.GetProfile)
				// userGroup.PATCH("/tier", userHandler.UpdateTier)
			}
		}
	}
	return router
}
