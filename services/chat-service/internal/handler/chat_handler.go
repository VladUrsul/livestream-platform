package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/hub"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS handled at gateway level
	},
}

type Handler struct {
	hub       *hub.Hub
	jwtSecret string
}

func New(h *hub.Hub, jwtSecret string) *Handler {
	return &Handler{hub: h, jwtSecret: jwtSecret}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/ws/:room", h.WebSocket)
	r.POST("/api/v1/chat/:room/slow-mode", h.SetSlowMode)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "chat-service"})
	})
}

// GET /ws/:room?token=<jwt>
func (h *Handler) WebSocket(c *gin.Context) {
	roomID := c.Param("room")

	// JWT from query param (WebSocket can't send headers)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := h.parseToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := hub.NewClient(h.hub, conn, roomID, claims.UserID, claims.Username)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// POST /api/v1/chat/:room/slow-mode  body: {"seconds": 10}
func (h *Handler) SetSlowMode(c *gin.Context) {
	roomID := c.Param("room")

	// Auth
	tokenStr := extractBearer(c)
	claims, err := h.parseToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Only the channel owner (roomID == username) can set slow mode
	if !strings.EqualFold(claims.Username, roomID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the channel owner can set slow mode"})
		return
	}

	var body struct {
		Seconds int `json:"seconds" binding:"min=0,max=120"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.hub.SetSlowMode(c.Request.Context(), roomID, body.Seconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slow_mode": body.Seconds})
}

// ── JWT ───────────────────────────────────────────────────────────

type jwtClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	jwt.RegisteredClaims
}

func (h *Handler) parseToken(tokenStr string) (*jwtClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return parsed.Claims.(*jwtClaims), nil
}

func extractBearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return h[7:]
	}
	return ""
}
