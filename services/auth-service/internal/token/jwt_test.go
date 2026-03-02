package token

import (
	"testing"
	"time"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

func newTestProvider() *Provider {
	return NewProvider("access-secret-test", "refresh-secret-test", 15*time.Minute, 7*24*time.Hour)
}

func newTestUser() *domain.User {
	return &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
	}
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	p := newTestProvider()
	user := newTestUser()

	tokenStr, err := p.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}

	claims, err := p.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("expected valid token, got: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected user_id %v, got %v", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		t.Errorf("expected email %v, got %v", user.Email, claims.Email)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	p := newTestProvider()
	user := newTestUser()

	tokenStr, err := p.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	claims, err := p.ValidateRefreshToken(tokenStr)
	if err != nil {
		t.Fatalf("expected valid refresh token, got: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected user_id %v, got %v", user.ID, claims.UserID)
	}
}

func TestAccessTokenCannotBeValidatedAsRefresh(t *testing.T) {
	p := newTestProvider()
	user := newTestUser()

	accessToken, _ := p.GenerateAccessToken(user)
	_, err := p.ValidateRefreshToken(accessToken)
	if err == nil {
		t.Fatal("expected error when validating access token as refresh token")
	}
}

func TestExpiredAccessToken(t *testing.T) {
	p := NewProvider("access-secret", "refresh-secret", -1*time.Minute, 7*24*time.Hour)
	user := newTestUser()

	tokenStr, _ := p.GenerateAccessToken(user)
	_, err := p.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestInvalidTokenString(t *testing.T) {
	p := newTestProvider()
	_, err := p.ValidateAccessToken("this.is.not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}
