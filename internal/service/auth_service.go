package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/Nutan-Kum12/RateLimiterX/internal/auth"
	"github.com/Nutan-Kum12/RateLimiterX/internal/dto"
	"github.com/Nutan-Kum12/RateLimiterX/internal/logger"
	"github.com/Nutan-Kum12/RateLimiterX/internal/model"
	"github.com/Nutan-Kum12/RateLimiterX/internal/repository"
)

// AuthService defines the interface for authentication business logic.
type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user with the "free" tier and returns JWT tokens.
func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user already exists
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	now := time.Now()
	user := &model.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Tier:      model.TierFree,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Log.Info("user registered successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("tier", user.Tier),
	)

	// Generate JWT tokens
	return s.generateTokens(user)
}

// Login validates credentials and returns JWT tokens.
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	logger.Log.Info("user logged in successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	)

	// Generate JWT tokens
	return s.generateTokens(user)
}

// generateTokens creates access and refresh tokens for a user.
func (s *authService) generateTokens(user *model.User) (*dto.AuthResponse, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Tier)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Tier:  user.Tier,
		},
	}, nil
}
