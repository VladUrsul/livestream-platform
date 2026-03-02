package handler

import (
	"errors"
	"net/http"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles HTTP requests for auth endpoints.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes mounts all auth routes onto the given router group.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", h.Register)
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.Refresh)
	rg.POST("/logout", h.Logout)
	rg.GET("/validate", h.ValidateToken)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input domain.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input domain.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user id"})
		return
	}

	accessToken := extractBearerToken(c)
	if err := h.authService.Logout(c.Request.Context(), userID, accessToken); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "logout failed"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "logged out successfully"})
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	accessToken := extractBearerToken(c)
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing token"})
		return
	}

	claims, err := h.authService.ValidateToken(c.Request.Context(), accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	c.JSON(http.StatusOK, claims)
}

func (h *AuthHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrEmailTaken):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrUsernameTaken):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrUserNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
