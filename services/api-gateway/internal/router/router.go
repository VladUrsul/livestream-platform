package router

import (
	"net/http"

	"github.com/VladUrsul/livestream-platform/services/api-gateway/internal/config"
	"github.com/VladUrsul/livestream-platform/services/api-gateway/internal/middleware"
	"github.com/VladUrsul/livestream-platform/services/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

// New builds and returns the gateway router with all routes registered.
func New(cfg *config.Config) (*gin.Engine, error) {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "api-gateway"})
	})

	// ── Build service proxies ──────────────────────────────────────────────
	authProxy, err := proxy.NewServiceProxy(cfg.Services.AuthServiceURL)
	if err != nil {
		return nil, err
	}

	streamProxy, err := proxy.NewServiceProxy(cfg.Services.StreamServiceURL)
	if err != nil {
		return nil, err
	}

	// ── Route groups ──────────────────────────────────────────────────────
	// Each group maps to one microservice.

	api := r.Group("/api/v1")

	// Auth routes — no authentication required, forwarded directly
	api.Any("/auth/*path", authProxy.Forward())
	api.Any("/streams/*path", streamProxy.Forward())

	// userProxy, _ := proxy.NewServiceProxy(cfg.Services.UserServiceURL)
	// api.Any("/users/*path", userProxy.Forward())

	// streamProxy, _ := proxy.NewServiceProxy(cfg.Services.StreamServiceURL)
	// api.Any("/streams/*path", streamProxy.Forward())

	return r, nil
}
