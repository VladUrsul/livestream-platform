package token

import (
	"fmt"
	"time"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Provider handles JWT creation and validation.
type Provider struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// jwtClaims extends jwt.RegisteredClaims with our custom fields.
type jwtClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// NewProvider creates a new JWT provider.
func NewProvider(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *Provider {
	return &Provider{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken creates a signed JWT access token for a user.
func (p *Provider) GenerateAccessToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(p.accessExpiry)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(p.accessSecret)
}

// GenerateRefreshToken creates a signed JWT refresh token.
func (p *Provider) GenerateRefreshToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(p.refreshExpiry)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(p.refreshSecret)
}

// ValidateAccessToken parses and validates an access token, returning its claims.
func (p *Provider) ValidateAccessToken(tokenStr string) (*domain.Claims, error) {
	return p.validate(tokenStr, p.accessSecret)
}

// ValidateRefreshToken parses and validates a refresh token, returning its claims.
func (p *Provider) ValidateRefreshToken(tokenStr string) (*domain.Claims, error) {
	return p.validate(tokenStr, p.refreshSecret)
}

// AccessExpirySeconds returns how many seconds until an access token expires.
func (p *Provider) AccessExpirySeconds() int64 {
	return int64(p.accessExpiry.Seconds())
}

// AccessExpiry returns the access token duration (used by cache layer).
func (p *Provider) AccessExpiry() time.Duration {
	return p.accessExpiry
}

// RefreshExpiry returns the refresh token duration (used by cache layer).
func (p *Provider) RefreshExpiry() time.Duration {
	return p.refreshExpiry
}

func (p *Provider) validate(tokenStr string, secret []byte) (*domain.Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &domain.Claims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
	}, nil
}
