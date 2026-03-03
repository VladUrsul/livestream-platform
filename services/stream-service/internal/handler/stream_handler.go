package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StreamHandler struct {
	svc service.StreamService
}

func NewStreamHandler(svc service.StreamService) *StreamHandler {
	return &StreamHandler{svc: svc}
}

// GET /api/v1/streams/key
func (h *StreamHandler) GetStreamKey(c *gin.Context) {
	userID, username, ok := extractCaller(c)
	if !ok {
		return
	}
	key, err := h.svc.GetOrCreateStreamKey(c.Request.Context(), userID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stream key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"stream_key": key.Key,
		"rtmp_url":   "rtmp://localhost:1935/live",
		"obs_url":    "rtmp://localhost:1935/live/" + key.Key,
	})
}

// POST /api/v1/streams/key/rotate
func (h *StreamHandler) RotateStreamKey(c *gin.Context) {
	userID, _, ok := extractCaller(c)
	if !ok {
		return
	}
	key, err := h.svc.RotateStreamKey(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"stream_key": key.Key,
		"rtmp_url":   "rtmp://localhost:1935/live",
		"obs_url":    "rtmp://localhost:1935/live/" + key.Key,
	})
}

// PUT /api/v1/streams/settings
func (h *StreamHandler) UpdateSettings(c *gin.Context) {
	userID, _, ok := extractCaller(c)
	if !ok {
		return
	}
	var input domain.UpdateStreamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateStreamSettings(c.Request.Context(), userID, input); err != nil {
		if errors.Is(err, service.ErrStreamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

// GET /api/v1/streams/live
func (h *StreamHandler) GetLiveStreams(c *gin.Context) {
	streams, err := h.svc.GetLiveStreams(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get live streams"})
		return
	}
	if streams == nil {
		streams = []*domain.StreamPublicInfo{}
	}
	c.JSON(http.StatusOK, gin.H{"streams": streams})
}

// GET /api/v1/streams/:username
func (h *StreamHandler) GetStreamInfo(c *gin.Context) {
	info, err := h.svc.GetStreamInfo(c.Request.Context(), c.Param("username"))
	if err != nil {
		if errors.Is(err, service.ErrStreamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stream"})
		return
	}
	c.JSON(http.StatusOK, info)
}

// POST /api/v1/streams/:username/join
func (h *StreamHandler) JoinStream(c *gin.Context) {
	count, err := h.svc.JoinStream(c.Request.Context(), c.Param("username"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"viewer_count": count})
}

// POST /api/v1/streams/:username/leave
func (h *StreamHandler) LeaveStream(c *gin.Context) {
	count, err := h.svc.LeaveStream(c.Request.Context(), c.Param("username"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"viewer_count": count})
}

func extractCaller(c *gin.Context) (uuid.UUID, string, bool) {
	rawID, _ := c.Get("user_id")
	rawUsername, _ := c.Get("username")
	userID, err := uuid.Parse(fmt.Sprintf("%v", rawID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, "", false
	}
	return userID, fmt.Sprintf("%v", rawUsername), true
}
