package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/configs"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/logger"
)

// NewConnection creates and configures a MySQL connection pool.
func NewConnection(cfg configs.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Verify connectivity with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	logger.Log.Info("MySQL connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
	)

	return db, nil
}

// RunMigrations executes SQL migration files against the database.
// It reads .sql files and executes them in order.
func RunMigrations(db *sql.DB, migrations []string) error {
	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	logger.Log.Info("database migrations completed successfully")
	return nil
}

// Close gracefully closes the database connection pool.
func Close(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			logger.Log.Error("failed to close MySQL connection", zap.Error(err))
		} else {
			logger.Log.Info("MySQL connection closed")
		}
	}
}
