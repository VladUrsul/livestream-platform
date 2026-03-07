package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/hub"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
	hub       *hub.Hub
	repo      repository.NotificationRepository
	jwtSecret string
}

func New(h *hub.Hub, repo repository.NotificationRepository, jwtSecret string) *Handler {
	return &Handler{hub: h, repo: repo, jwtSecret: jwtSecret}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "notification-service"})
	})
	r.GET("/ws/notifications", h.WebSocket)

	api := r.Group("/api/v1/notifications")
	api.Use(h.authMiddleware())
	api.GET("", h.GetNotifications)
	api.PUT("/read-all", h.MarkAllRead)
	api.PUT("/:id/read", h.MarkRead)
}

// GET /ws/notifications?token=<jwt>
func (h *Handler) WebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	claims, err := hub.ParseToken(tokenStr, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := h.hub.Connect(claims.UserID, conn)

	// Send unread count immediately on connect
	unread, _ := h.repo.GetUnreadCount(c.Request.Context(), claims.UserID)
	data, _ := marshalPush(&domain.WSPush{Type: "unread_count", UnreadCount: unread})
	client.GetSend() <- data

	go h.hub.WritePump(client)
	h.hub.ReadPump(client, func() {
		h.hub.Disconnect(claims.UserID)
	})
}

// GET /api/v1/notifications?limit=20&offset=0
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	notifications, err := h.repo.GetForUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	if notifications == nil {
		notifications = []*domain.Notification{}
	}
	unread, _ := h.repo.GetUnreadCount(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"notifications": notifications, "unread_count": unread})
}

// PUT /api/v1/notifications/read-all
func (h *Handler) MarkAllRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.repo.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PUT /api/v1/notifications/:id/read
func (h *Handler) MarkRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.repo.MarkRead(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		claims, err := hub.ParseToken(auth[7:], h.jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func marshalPush(p *domain.WSPush) ([]byte, error) {

	return json.Marshal(p)
}
