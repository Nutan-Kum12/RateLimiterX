package model

import "time"

type User struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Tier      string    `json:"tier" db:"tier"` // "free" | "premium"
}

// Tier constants for user classification.
const (
	TierFree    = "free"
	TierPremium = "premium"
)

// ValidTiers contains all valid tier values.
var ValidTiers = map[string]bool{
	TierFree:    true,
	TierPremium: true,
}

// IsValidTier checks if the given tier string is valid.
func IsValidTier(tier string) bool {
	return ValidTiers[tier]
}
