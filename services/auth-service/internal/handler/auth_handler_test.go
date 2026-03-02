package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockAuthService struct {
	registerFn func(ctx context.Context, input domain.RegisterInput) (*domain.AuthResponse, error)
	loginFn    func(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error)
	refreshFn  func(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
	logoutFn   func(ctx context.Context, userID uuid.UUID, accessToken string) error
	validateFn func(ctx context.Context, accessToken string) (*domain.Claims, error)
}

func (m *mockAuthService) Register(ctx context.Context, input domain.RegisterInput) (*domain.AuthResponse, error) {
	return m.registerFn(ctx, input)
}
func (m *mockAuthService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error) {
	return m.loginFn(ctx, input)
}
func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	return m.refreshFn(ctx, refreshToken)
}
func (m *mockAuthService) Logout(ctx context.Context, userID uuid.UUID, accessToken string) error {
	return m.logoutFn(ctx, userID, accessToken)
}
func (m *mockAuthService) ValidateToken(ctx context.Context, accessToken string) (*domain.Claims, error) {
	return m.validateFn(ctx, accessToken)
}

func newTestRouter(svc service.AuthService) *gin.Engine {
	r := gin.New()
	h := NewAuthHandler(svc)
	h.RegisterRoutes(r.Group("/auth"))
	return r
}

func fakeAuthResponse() *domain.AuthResponse {
	return &domain.AuthResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
		User: domain.PublicUser{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Username:  "testuser",
			CreatedAt: time.Now(),
		},
	}
}

func TestRegisterHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _ domain.RegisterInput) (*domain.AuthResponse, error) {
			return fakeAuthResponse(), nil
		},
	}

	body, _ := json.Marshal(map[string]string{
		"email": "test@example.com", "username": "testuser", "password": "password123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(svc).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(&mockAuthService{}).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegisterHandler_EmailConflict(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _ domain.RegisterInput) (*domain.AuthResponse, error) {
			return nil, service.ErrEmailTaken
		},
	}

	body, _ := json.Marshal(map[string]string{
		"email": "taken@example.com", "username": "user", "password": "pass12345",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(svc).ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, _ domain.LoginInput) (*domain.AuthResponse, error) {
			return fakeAuthResponse(), nil
		},
	}

	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "pass"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(svc).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, _ domain.LoginInput) (*domain.AuthResponse, error) {
			return nil, service.ErrInvalidCredentials
		},
	}

	body, _ := json.Marshal(map[string]string{"email": "x@x.com", "password": "wrong"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(svc).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRefreshHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		refreshFn: func(_ context.Context, _ string) (*domain.AuthResponse, error) {
			return fakeAuthResponse(), nil
		},
	}

	body, _ := json.Marshal(map[string]string{"refresh_token": "valid-token"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(svc).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestValidateHandler_InvalidToken(t *testing.T) {
	svc := &mockAuthService{
		validateFn: func(_ context.Context, _ string) (*domain.Claims, error) {
			return nil, errors.New("invalid")
		},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	newTestRouter(svc).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
