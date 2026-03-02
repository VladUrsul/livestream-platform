package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/cache"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/repository"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors for consistent error handling across layers.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// AuthService defines the auth business logic contract.
type AuthService interface {
	Register(ctx context.Context, input domain.RegisterInput) (*domain.AuthResponse, error)
	Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, accessToken string) error
	ValidateToken(ctx context.Context, accessToken string) (*domain.Claims, error)
}

type authService struct {
	userRepo      repository.UserRepository
	tokenProvider *token.Provider
	authCache     cache.AuthCache
}

// NewAuthService creates a new auth service with all dependencies injected.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenProvider *token.Provider,
	authCache cache.AuthCache,
) AuthService {
	return &authService{
		userRepo:      userRepo,
		tokenProvider: tokenProvider,
		authCache:     authCache,
	}
}

// Register creates a new user account and returns auth tokens.
func (s *authService) Register(ctx context.Context, input domain.RegisterInput) (*domain.AuthResponse, error) {
	emailExists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("register: check email: %w", err)
	}
	if emailExists {
		return nil, ErrEmailTaken
	}

	usernameExists, err := s.userRepo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("register: check username: %w", err)
	}
	if usernameExists {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("register: hash password: %w", err)
	}

	user := &domain.User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, ErrEmailTaken
		}
		if errors.Is(err, repository.ErrDuplicateUsername) {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("register: create user: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

// Login authenticates a user and returns auth tokens.
func (s *authService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateAuthResponse(ctx, user)
}

// Refresh exchanges a valid refresh token for new access + refresh tokens.
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	claims, err := s.tokenProvider.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("refresh: find user: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

// Logout invalidates the user's current access token.
func (s *authService) Logout(ctx context.Context, userID uuid.UUID, accessToken string) error {
	if err := s.authCache.BlacklistToken(ctx, accessToken, s.tokenProvider.AccessExpiry()); err != nil {
		return fmt.Errorf("logout: blacklist token: %w", err)
	}
	if err := s.authCache.DeleteRefreshToken(ctx, userID.String()); err != nil {
		return fmt.Errorf("logout: delete refresh token: %w", err)
	}
	return nil
}

// ValidateToken checks an access token is valid and not blacklisted.
func (s *authService) ValidateToken(ctx context.Context, accessToken string) (*domain.Claims, error) {
	blacklisted, err := s.authCache.IsBlacklisted(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("validate token: check blacklist: %w", err)
	}
	if blacklisted {
		return nil, ErrInvalidToken
	}

	claims, err := s.tokenProvider.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *authService) generateAuthResponse(ctx context.Context, user *domain.User) (*domain.AuthResponse, error) {
	accessToken, err := s.tokenProvider.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenProvider.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.authCache.StoreRefreshToken(
		ctx,
		user.ID.String(),
		refreshToken,
		s.tokenProvider.RefreshExpiry(),
	); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.tokenProvider.AccessExpirySeconds(),
		User:         user.ToPublic(),
	}, nil
}
