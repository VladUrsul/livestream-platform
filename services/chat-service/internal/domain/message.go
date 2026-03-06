package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	RoomID    string    `json:"room_id"` // = channel username
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Room struct {
	ID        string    `json:"id"`        // = channel username
	SlowMode  int       `json:"slow_mode"` // seconds between messages, 0 = off
	CreatedAt time.Time `json:"created_at"`
}

type WSMessage struct {
	Type     string    `json:"type"` // "message" | "history" | "error" | "slow_mode"
	Message  *Message  `json:"message,omitempty"`
	History  []Message `json:"history,omitempty"`
	Error    string    `json:"error,omitempty"`
	SlowMode int       `json:"slow_mode,omitempty"`
}
