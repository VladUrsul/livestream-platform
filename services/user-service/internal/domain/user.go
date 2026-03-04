package domain

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	AvatarURL   string    `json:"avatar_url"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	IsLive      bool      `json:"is_live"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateProfileInput struct {
	DisplayName string `json:"display_name" binding:"max=60"`
	Bio         string `json:"bio"          binding:"max=200"`
	AvatarURL   string `json:"avatar_url"   binding:"omitempty,url"`
}

type UserRegisteredEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchResult struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Followers   int       `json:"followers"`
	IsLive      bool      `json:"is_live"`
}
