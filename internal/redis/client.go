package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/Nutan-Kum12/RateLimiterX/internal/configs"
	"github.com/Nutan-Kum12/RateLimiterX/internal/logger"
)

// NewClient creates and validates a new Redis client connection.
func NewClient(cfg configs.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     50,
		MinIdleConns: 10,
	})

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Log.Info("Redis connected successfully",
		zap.String("addr", cfg.Addr),
		zap.Int("db", cfg.DB),
	)

	return client, nil
}

// Close gracefully closes the Redis client.
func Close(client *redis.Client) {
	if client != nil {
		if err := client.Close(); err != nil {
			logger.Log.Error("failed to close Redis connection", zap.Error(err))
		} else {
			logger.Log.Info("Redis connection closed")
		}
	}
}
