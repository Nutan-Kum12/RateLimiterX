package service

import (
	"context"
	"fmt"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/dto"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/repository"
)

// UserService defines the interface for user management business logic.
type UserService interface {
	GetProfile(ctx context.Context, userID string) (*dto.UserInfo, error)
	// UpdateTier(ctx context.Context, userID string, tier string) (*dto.UserInfo, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService instance.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetProfile retrieves the public profile of a user.
func (s *userService) GetProfile(ctx context.Context, userID string) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &dto.UserInfo{
		ID:    user.ID,
		Email: user.Email,
		Tier:  user.Tier,
	}, nil
}

// UpdateTier changes a user's tier (e.g., free → premium).
// func (s *userService) UpdateTier(ctx context.Context, userID string, tier string) (*dto.UserInfo, error) {
// 	// Validate tier value
// 	if !model.IsValidTier(tier) {
// 		return nil, fmt.Errorf("invalid tier: %s", tier)
// 	}
// 	// Update tier in the database
// 	if err := s.userRepo.UpdateTier(ctx, userID, tier); err != nil {
// 		return nil, fmt.Errorf("failed to update tier: %w", err)
// 	}
// 	logger.Log.Info("user tier updated",
// 		zap.String("user_id", userID),
// 		zap.String("new_tier", tier),
// 	)
// 	// Return updated profile
// 	return s.GetProfile(ctx, userID)
// }