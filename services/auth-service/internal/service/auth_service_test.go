package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/repository"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/token"
	"github.com/google/uuid"
)

// ---- Mock UserRepository ----

type mockUserRepo struct {
	users          map[string]*domain.User
	emailExists    bool
	usernameExists bool
	createErr      error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = uuid.New()
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return m.emailExists, nil
}

func (m *mockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return m.usernameExists, nil
}

// ---- Mock AuthCache ----

type mockAuthCache struct {
	blacklisted   map[string]bool
	refreshTokens map[string]string
}

func newMockAuthCache() *mockAuthCache {
	return &mockAuthCache{
		blacklisted:   make(map[string]bool),
		refreshTokens: make(map[string]string),
	}
}

func (m *mockAuthCache) BlacklistToken(ctx context.Context, tokenID string, expiry time.Duration) error {
	m.blacklisted[tokenID] = true
	return nil
}

func (m *mockAuthCache) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	return m.blacklisted[tokenID], nil
}

func (m *mockAuthCache) StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiry time.Duration) error {
	m.refreshTokens[userID] = tokenHash
	return nil
}

func (m *mockAuthCache) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	return m.refreshTokens[userID], nil
}

func (m *mockAuthCache) DeleteRefreshToken(ctx context.Context, userID string) error {
	delete(m.refreshTokens, userID)
	return nil
}

// ---- Test helper ----

func newTestService(repo *mockUserRepo, c *mockAuthCache) AuthService {
	p := token.NewProvider("access-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)
	return NewAuthService(repo, p, c)
}

// ---- Tests ----

func TestRegister_Success(t *testing.T) {
	svc := newTestService(newMockUserRepo(), newMockAuthCache())

	resp, err := svc.Register(context.Background(), domain.RegisterInput{
		Email:    "user@example.com",
		Username: "testuser",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.User.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", resp.User.Email)
	}
}

func TestRegister_EmailAlreadyTaken(t *testing.T) {
	repo := newMockUserRepo()
	repo.emailExists = true
	svc := newTestService(repo, newMockAuthCache())

	_, err := svc.Register(context.Background(), domain.RegisterInput{
		Email: "taken@example.com", Username: "user", Password: "password123",
	})

	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got: %v", err)
	}
}

func TestRegister_UsernameTaken(t *testing.T) {
	repo := newMockUserRepo()
	repo.usernameExists = true
	svc := newTestService(repo, newMockAuthCache())

	_, err := svc.Register(context.Background(), domain.RegisterInput{
		Email: "new@example.com", Username: "takenuser", Password: "password123",
	})

	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := newTestService(repo, newMockAuthCache())
	ctx := context.Background()

	svc.Register(ctx, domain.RegisterInput{
		Email: "login@example.com", Username: "loginuser", Password: "mypassword",
	})

	resp, err := svc.Login(ctx, domain.LoginInput{
		Email: "login@example.com", Password: "mypassword",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := newTestService(repo, newMockAuthCache())
	ctx := context.Background()

	svc.Register(ctx, domain.RegisterInput{
		Email: "user@example.com", Username: "user", Password: "correctpassword",
	})

	_, err := svc.Login(ctx, domain.LoginInput{
		Email: "user@example.com", Password: "wrongpassword",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := newTestService(newMockUserRepo(), newMockAuthCache())

	_, err := svc.Login(context.Background(), domain.LoginInput{
		Email: "ghost@example.com", Password: "whatever",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogout_BlacklistsToken(t *testing.T) {
	c := newMockAuthCache()
	svc := newTestService(newMockUserRepo(), c)
	ctx := context.Background()

	resp, _ := svc.Register(ctx, domain.RegisterInput{
		Email: "u@example.com", Username: "u", Password: "password123",
	})

	err := svc.Logout(ctx, resp.User.ID, resp.AccessToken)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !c.blacklisted[resp.AccessToken] {
		t.Error("expected access token to be blacklisted after logout")
	}
}

func TestValidateToken_BlacklistedToken(t *testing.T) {
	c := newMockAuthCache()
	svc := newTestService(newMockUserRepo(), c)
	ctx := context.Background()

	resp, _ := svc.Register(ctx, domain.RegisterInput{
		Email: "v@example.com", Username: "v", Password: "password123",
	})

	svc.Logout(ctx, resp.User.ID, resp.AccessToken)

	_, err := svc.ValidateToken(ctx, resp.AccessToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for blacklisted token, got: %v", err)
	}
}
