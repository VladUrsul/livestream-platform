package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifyFollowed   NotificationType = "user.followed"
	NotifyStreamLive NotificationType = "stream.started"
)

type Notification struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"` // recipient
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	ActorID   uuid.UUID        `json:"actor_id"`
	ActorName string           `json:"actor_name"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"created_at"`
}

// Events consumed from RabbitMQ
type UserFollowedEvent struct {
	FollowerID       uuid.UUID `json:"follower_id"`
	FollowerUsername string    `json:"follower_username"`
	FolloweeID       uuid.UUID `json:"followee_id"`
}

type StreamStartedEvent struct {
	EventType string    `json:"event_type"`
	StreamID  uuid.UUID `json:"stream_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Title     string    `json:"title"`
}

// WSPush is sent over WebSocket to the client
type WSPush struct {
	Type         string        `json:"type"` // "notification" | "unread_count"
	Notification *Notification `json:"notification,omitempty"`
	UnreadCount  int           `json:"unread_count,omitempty"`
}
