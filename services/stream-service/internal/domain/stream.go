package domain

import (
	"time"

	"github.com/google/uuid"
)

type StreamStatus string

const (
	StreamStatusOffline StreamStatus = "offline"
	StreamStatusLive    StreamStatus = "live"
	StreamStatusEnded   StreamStatus = "ended"
)

type Stream struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	Username    string       `json:"username"`
	Title       string       `json:"title"`
	Category    string       `json:"category"`
	StreamKey   string       `json:"stream_key,omitempty"`
	Status      StreamStatus `json:"status"`
	ViewerCount int          `json:"viewer_count"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	EndedAt     *time.Time   `json:"ended_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type StreamKey struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateStreamInput struct {
	Title    string `json:"title"    binding:"required,min=3,max=120"`
	Category string `json:"category" binding:"required"`
}

type StreamPublicInfo struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	Username    string       `json:"username"`
	Title       string       `json:"title"`
	Category    string       `json:"category"`
	Status      StreamStatus `json:"status"`
	ViewerCount int          `json:"viewer_count"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	HLSUrl      string       `json:"hls_url,omitempty"`
}

func (s *Stream) ToPublic(hlsBaseURL string) StreamPublicInfo {
	info := StreamPublicInfo{
		ID:          s.ID,
		UserID:      s.UserID,
		Username:    s.Username,
		Title:       s.Title,
		Category:    s.Category,
		Status:      s.Status,
		ViewerCount: s.ViewerCount,
		StartedAt:   s.StartedAt,
	}
	if s.Status == StreamStatusLive {
		info.HLSUrl = hlsBaseURL + "/" + s.Username + "/index.m3u8"
	}
	return info
}

type StreamStartedEvent struct {
	EventType string    `json:"event_type"`
	StreamID  uuid.UUID `json:"stream_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Title     string    `json:"title"`
	StartedAt time.Time `json:"started_at"`
}

type StreamEndedEvent struct {
	EventType string    `json:"event_type"`
	StreamID  uuid.UUID `json:"stream_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	EndedAt   time.Time `json:"ended_at"`
}
