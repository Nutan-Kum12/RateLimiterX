package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // pprof is intentionally exposed on the dedicated profiling server
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Nutan-Kum12/RateLimiterX/internal/api"
	"github.com/Nutan-Kum12/RateLimiterX/internal/auth"
	"github.com/Nutan-Kum12/RateLimiterX/internal/configs"
	"github.com/Nutan-Kum12/RateLimiterX/internal/limiter"
	"github.com/Nutan-Kum12/RateLimiterX/internal/logger"
	internalMySQL "github.com/Nutan-Kum12/RateLimiterX/internal/mysql"
	internalRedis "github.com/Nutan-Kum12/RateLimiterX/internal/redis"
	"github.com/Nutan-Kum12/RateLimiterX/internal/repository"
	"github.com/Nutan-Kum12/RateLimiterX/internal/service"
)

// SQL migrations to run on startup.
var migrations = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id         VARCHAR(36) PRIMARY KEY,
		email      VARCHAR(255) UNIQUE NOT NULL,
		password   VARCHAR(255) NOT NULL,
		tier       ENUM('free', 'premium') DEFAULT 'free',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_email (email)
	);`,
}

func main() {

	// Load Configuration
	cfg, err := configs.Load("config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize Logger
	logger.Init(cfg.Server.Mode)
	defer logger.Sync()

	logger.Log.Info("starting RateLimiterX",
		zap.Int("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode),
	)

	//  Connect to MySQL
	db, err := internalMySQL.NewConnection(cfg.Database)
	if err != nil {
		logger.Log.Fatal("failed to connect to MySQL", zap.Error(err))
	}
	defer internalMySQL.Close(db)

	// Run migrations
	if err := internalMySQL.RunMigrations(db, migrations); err != nil {
		logger.Log.Fatal("failed to run migrations", zap.Error(err))
	}

	// Connect to Redis
	redisClient, err := internalRedis.NewClient(cfg.Redis)
	if err != nil {
		logger.Log.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer internalRedis.Close(redisClient)

	//  Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	policyRepo := repository.NewPolicyRepository(cfg.Tiers)

	// Initialize JWT Manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)
	// Initialize Service
	authService := service.NewAuthService(userRepo, jwtManager)
	userService := service.NewUserService(userRepo)

	// Initialize Rate Limiter
	limiterManager, err := limiter.NewManager(redisClient, policyRepo)
	if err != nil {
		logger.Log.Fatal("failed to initialize rate limiter", zap.Error(err))
	}

	//  Setup Router
	router := api.NewRouter(cfg.Server.Mode, api.RouterDeps{
		AuthService:    authService,
		UserService:    userService,
		JWTManager:     jwtManager,
		LimiterManager: limiterManager,
		DB:             db,
		RedisClient:    redisClient,
	})

	// Start HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	// pprof server
	pprofSrv := &http.Server{
		Addr:              ":6060",
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Log.Info("HTTP server starting",
			zap.String("addr", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("HTTP server error", zap.Error(err))
		}
	}()
	// Start pprof server
	go func() {
		logger.Log.Info("pprof server starting",
			zap.String("addr", pprofSrv.Addr),
		)

		if err := pprofSrv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			logger.Log.Error("pprof server error", zap.Error(err))
		}
	}()
	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("server exited gracefully")
}
