package dto

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned after successful registration or login.
type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

// UserInfo is the public-facing user representation (no password).
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Tier  string `json:"tier"`
}

// UpdateTierRequest is the payload for updating a user's tier.
type UpdateTierRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Tier   string `json:"tier" binding:"required,oneof=free premium"`
}
