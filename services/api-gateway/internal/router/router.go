package router

import (
	"net/http"

	"github.com/VladUrsul/livestream-platform/services/api-gateway/internal/config"
	"github.com/VladUrsul/livestream-platform/services/api-gateway/internal/middleware"
	"github.com/VladUrsul/livestream-platform/services/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config) (*gin.Engine, error) {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "api-gateway"})
	})

	authProxy, err := proxy.NewServiceProxy(cfg.Services.AuthServiceURL)
	if err != nil {
		return nil, err
	}
	streamProxy, err := proxy.NewServiceProxy(cfg.Services.StreamServiceURL)
	if err != nil {
		return nil, err
	}
	userProxy, _ := proxy.NewServiceProxy(cfg.Services.UserServiceURL)
	chatProxy, _ := proxy.NewServiceProxy(cfg.Services.ChatServiceURL)
	notifProxy, _ := proxy.NewServiceProxy(cfg.Services.NotificationServiceURL)

	r.Any("/ws/*path", chatProxy.ForwardWS(notifProxy))

	r.Any("/api/v1/chat/*path", chatProxy.Forward())
	r.Any("/api/v1/notifications/*path", notifProxy.Forward())

	api := r.Group("/api/v1")
	api.Any("/auth/*path", authProxy.Forward())
	api.Any("/streams/*path", streamProxy.Forward())
	api.Any("/users/*path", userProxy.Forward())

	return r, nil
}
