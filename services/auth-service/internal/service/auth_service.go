package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/cache"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/repository"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// EventPublisher is satisfied by publisher.Publisher — defined here to avoid
// an import cycle and to keep the service layer decoupled from infrastructure.
type EventPublisher interface {
	Publish(routingKey string, payload any) error
}

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
	publisher     EventPublisher // nil = no events (RabbitMQ unavailable)
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenProvider *token.Provider,
	authCache cache.AuthCache,
	pub EventPublisher,
) AuthService {
	return &authService{
		userRepo:      userRepo,
		tokenProvider: tokenProvider,
		authCache:     authCache,
		publisher:     pub,
	}
}

// SetPublisher wires in the event publisher after construction.
// Called from main.go once RabbitMQ is available.
func (s *authService) SetPublisher(p EventPublisher) {
	s.publisher = p
}

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

	// Publish user.registered so user-service creates the profile
	if s.publisher != nil {
		_ = s.publisher.Publish("user.registered", map[string]any{
			"user_id":    user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		})
		log.Printf("[Auth] published user.registered for @%s", user.Username)
	}

	return s.generateAuthResponse(ctx, user)
}

// Login, Refresh, Logout, ValidateToken, generateAuthResponse unchanged
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

func (s *authService) Logout(ctx context.Context, userID uuid.UUID, accessToken string) error {
	if err := s.authCache.BlacklistToken(ctx, accessToken, s.tokenProvider.AccessExpiry()); err != nil {
		return fmt.Errorf("logout: blacklist token: %w", err)
	}
	if err := s.authCache.DeleteRefreshToken(ctx, userID.String()); err != nil {
		return fmt.Errorf("logout: delete refresh token: %w", err)
	}
	return nil
}

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
	if err := s.authCache.StoreRefreshToken(ctx, user.ID.String(), refreshToken, s.tokenProvider.RefreshExpiry()); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}
	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.tokenProvider.AccessExpirySeconds(),
		User:         user.ToPublic(),
	}, nil
}
