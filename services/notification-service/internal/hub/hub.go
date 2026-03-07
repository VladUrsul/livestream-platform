package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	UserID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]*Client // userID → client
}

func New() *Hub {
	return &Hub{clients: make(map[uuid.UUID]*Client)}
}

func (c *Client) GetSend() chan []byte { return c.send }

func (h *Hub) Connect(userID uuid.UUID, conn *websocket.Conn) *Client {
	c := &Client{
		UserID: userID,
		conn:   conn,
		send:   make(chan []byte, 64),
	}
	h.mu.Lock()
	h.clients[userID] = c
	h.mu.Unlock()
	log.Printf("[Hub] user %s connected (total=%d)", userID, len(h.clients))
	return c
}

func (h *Hub) Disconnect(userID uuid.UUID) {
	h.mu.Lock()
	if c, ok := h.clients[userID]; ok {
		close(c.send)
		delete(h.clients, userID)
	}
	h.mu.Unlock()
	log.Printf("[Hub] user %s disconnected", userID)
}

func (h *Hub) Push(userID uuid.UUID, push *domain.WSPush) {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	data, err := json.Marshal(push)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		// Client too slow — disconnect
		h.Disconnect(userID)
	}
}

func (h *Hub) WritePump(c *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) ReadPump(c *Client, onClose func()) {
	defer func() {
		onClose()
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// JWT parsing (same approach as chat-service)
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func ParseToken(tokenStr, secret string) (*JWTClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, err
	}
	return parsed.Claims.(*JWTClaims), nil
}
