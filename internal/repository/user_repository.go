package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Nutan-Kum12/RateLimiterX/internal/model"
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	// UpdateTier(ctx context.Context, id string, tier string) error
}

// mysqlUserRepository implements UserRepository using MySQL.
type mysqlUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new MySQL-backed UserRepository.
// Creates a new repository instance and injects the shared *sql.DB.
func NewUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{db: db}
}

// Create inserts a new user into the database.
func (r *mysqlUserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, email, password, tier, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Password, user.Tier, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByEmail retrieves a user by their email address.
func (r *mysqlUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password, tier, created_at, updated_at
	          FROM users WHERE email = ?`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Tier, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found — not an error
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return user, nil
}

// FindByID retrieves a user by their unique ID.
func (r *mysqlUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, email, password, tier, created_at, updated_at
	         FROM users WHERE id = ?`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Tier, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return user, nil
}

// UpdateTier changes the tier for a specific user.
// func (r *mysqlUserRepository) UpdateTier(ctx context.Context, id string, tier string) error {
// 	query := `UPDATE users SET tier = ?, updated_at = NOW()
// 	          WHERE id = ?`

// 	result, err := r.db.ExecContext(ctx, query, tier, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to update user tier: %w", err)
// 	}

// 	rows, err := result.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("failed to check affected rows: %w", err)
// 	}
// 	if rows == 0 {
// 		return fmt.Errorf("user not found with ID: %s", id)
// 	}
// 	return nil
// }
