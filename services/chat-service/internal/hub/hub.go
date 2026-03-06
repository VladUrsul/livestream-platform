package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/repository"
)

type IncomingMessage struct {
	Client *Client
	Data   []byte
}

type Hub struct {
	rooms        map[string]map[*Client]bool // roomID → set of clients
	register     chan *Client
	unregister   chan *Client
	incoming     chan *IncomingMessage
	repo         repository.ChatRepository
	historyLimit int
	maxMsgLen    int
}

func NewHub(repo repository.ChatRepository, historyLimit, maxMsgLen int) *Hub {
	return &Hub{
		rooms:        make(map[string]map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		incoming:     make(chan *IncomingMessage, 256),
		repo:         repo,
		historyLimit: historyLimit,
		maxMsgLen:    maxMsgLen,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case c := <-h.register:
			if h.rooms[c.RoomID] == nil {
				h.rooms[c.RoomID] = make(map[*Client]bool)
				// Ensure room exists in DB
				h.repo.UpsertRoom(ctx, &domain.Room{ID: c.RoomID})
			}
			h.rooms[c.RoomID][c] = true
			log.Printf("[Hub] @%s joined room=%s (total=%d)", c.Username, c.RoomID, len(h.rooms[c.RoomID]))

			// Send history to the new client
			go h.sendHistory(ctx, c)

			// Send current slow mode setting
			go h.sendSlowMode(ctx, c)

		case c := <-h.unregister:
			if clients, ok := h.rooms[c.RoomID]; ok {
				if _, ok := clients[c]; ok {
					delete(clients, c)
					close(c.send)
					log.Printf("[Hub] @%s left room=%s (total=%d)", c.Username, c.RoomID, len(clients))
				}
				if len(clients) == 0 {
					delete(h.rooms, c.RoomID)
				}
			}

		case inc := <-h.incoming:
			h.handleIncoming(ctx, inc)
		}
	}
}

func (h *Hub) handleIncoming(ctx context.Context, inc *IncomingMessage) {
	c := inc.Client

	// Parse raw text as message content
	content := strings.TrimSpace(string(inc.Data))
	if content == "" {
		return
	}

	// Check max length
	if len(content) > h.maxMsgLen {
		h.sendError(c, "Message too long")
		return
	}

	// Slow mode check
	room, err := h.repo.GetRoom(ctx, c.RoomID)
	if err == nil && room.SlowMode > 0 {
		elapsed := time.Since(c.lastMsg).Seconds()
		if elapsed < float64(room.SlowMode) {
			remaining := float64(room.SlowMode) - elapsed
			h.sendError(c, "Slow mode: wait "+formatSeconds(remaining))
			return
		}
	}

	c.lastMsg = time.Now()

	msg := &domain.Message{
		RoomID:   c.RoomID,
		UserID:   c.UserID,
		Username: c.Username,
		Content:  content,
	}

	if err := h.repo.SaveMessage(ctx, msg); err != nil {
		log.Printf("[Hub] save message error: %v", err)
		return
	}

	h.broadcast(c.RoomID, &domain.WSMessage{
		Type:    "message",
		Message: msg,
	})
}

func (h *Hub) broadcast(roomID string, wsMsg *domain.WSMessage) {
	data, err := json.Marshal(wsMsg)
	if err != nil {
		return
	}
	for c := range h.rooms[roomID] {
		select {
		case c.send <- data:
		default:
			close(c.send)
			delete(h.rooms[roomID], c)
		}
	}
}

func (h *Hub) sendHistory(ctx context.Context, c *Client) {
	msgs, err := h.repo.GetHistory(ctx, c.RoomID, h.historyLimit)
	if err != nil {
		log.Printf("[Hub] get history error: %v", err)
		return
	}
	if msgs == nil {
		msgs = []domain.Message{}
	}
	data, _ := json.Marshal(&domain.WSMessage{Type: "history", History: msgs})
	select {
	case c.send <- data:
	default:
	}
}

func (h *Hub) sendSlowMode(ctx context.Context, c *Client) {
	room, err := h.repo.GetRoom(ctx, c.RoomID)
	if err != nil {
		return
	}
	data, _ := json.Marshal(&domain.WSMessage{Type: "slow_mode", SlowMode: room.SlowMode})
	select {
	case c.send <- data:
	default:
	}
}

func (h *Hub) sendError(c *Client, msg string) {
	data, _ := json.Marshal(&domain.WSMessage{Type: "error", Error: msg})
	select {
	case c.send <- data:
	default:
	}
}

// SetSlowMode is called from the HTTP handler for the channel owner
func (h *Hub) SetSlowMode(ctx context.Context, roomID string, seconds int) error {
	if err := h.repo.SetSlowMode(ctx, roomID, seconds); err != nil {
		return err
	}
	// Notify all clients in room
	h.broadcast(roomID, &domain.WSMessage{Type: "slow_mode", SlowMode: seconds})
	return nil
}

func formatSeconds(s float64) string {
	if s < 2 {
		return "1 second"
	}
	return strings.TrimRight(strings.TrimRight(
		strings.Replace(fmt.Sprintf("%.1f", s), ".", ".", 1),
		"0"), ".") + " seconds"
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}
