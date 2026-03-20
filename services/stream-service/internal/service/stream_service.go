package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/cache"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/hls"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/publisher"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrStreamNotFound   = errors.New("stream not found")
	ErrInvalidStreamKey = errors.New("invalid stream key")
	ErrAlreadyLive      = errors.New("stream already live")
)

type StreamService interface {
	GetOrCreateStreamKey(ctx context.Context, userID uuid.UUID, username string) (*domain.StreamKey, error)
	RotateStreamKey(ctx context.Context, userID uuid.UUID) (*domain.StreamKey, error)
	GetStreamInfo(ctx context.Context, username string) (*domain.StreamPublicInfo, error)
	GetLiveStreams(ctx context.Context) ([]*domain.StreamPublicInfo, error)
	UpdateStreamSettings(ctx context.Context, userID uuid.UUID, input domain.UpdateStreamInput) error

	HandleStreamStart(ctx context.Context, streamKey string) (io.WriteCloser, error)
	HandleStreamEnd(ctx context.Context, streamKey string)
	JoinStream(ctx context.Context, username string) (int, error)
	LeaveStream(ctx context.Context, username string) (int, error)
}

type streamService struct {
	repo        repository.StreamRepository
	streamCache cache.StreamCache
	transcoder  *hls.Transcoder
	pub         *publisher.StreamPublisher
	hlsBaseURL  string
}

func NewStreamService(
	repo repository.StreamRepository,
	streamCache cache.StreamCache,
	transcoder *hls.Transcoder,
	pub *publisher.StreamPublisher,
	hlsBaseURL string,
) StreamService {
	return &streamService{
		repo:        repo,
		streamCache: streamCache,
		transcoder:  transcoder,
		pub:         pub,
		hlsBaseURL:  hlsBaseURL,
	}
}

func (s *streamService) GetOrCreateStreamKey(ctx context.Context, userID uuid.UUID, username string) (*domain.StreamKey, error) {
	key, err := s.repo.GetStreamKeyByUserID(ctx, userID)
	if err == nil {
		return key, nil
	}
	if !errors.Is(err, repository.ErrKeyNotFound) {
		return nil, fmt.Errorf("get stream key: %w", err)
	}
	newKey := &domain.StreamKey{UserID: userID, Key: newStreamKey()}
	if err := s.repo.CreateStreamKey(ctx, newKey); err != nil {
		return nil, fmt.Errorf("create stream key: %w", err)
	}
	stream := &domain.Stream{
		UserID:    userID,
		Username:  username,
		Title:     username + "'s stream",
		Category:  "General",
		StreamKey: newKey.Key,
		Status:    domain.StreamStatusOffline,
	}
	if err := s.repo.CreateStream(ctx, stream); err != nil {
		return nil, fmt.Errorf("create stream record: %w", err)
	}
	return newKey, nil
}

func (s *streamService) RotateStreamKey(ctx context.Context, userID uuid.UUID) (*domain.StreamKey, error) {
	return s.repo.RotateStreamKey(ctx, userID)
}

func (s *streamService) GetStreamInfo(ctx context.Context, username string) (*domain.StreamPublicInfo, error) {
	stream, err := s.repo.GetStreamByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, fmt.Errorf("get stream info: %w", err)
	}
	// Reconcile stale "live" state:
	// if DB says live but transcoder isn't running in this pod, mark as ended.
	if stream.Status == domain.StreamStatusLive && !s.transcoder.IsRunning(stream.Username) {
		now := time.Now()
		stream.Status = domain.StreamStatusEnded
		stream.EndedAt = &now
		_ = s.streamCache.SetOffline(ctx, stream.Username)
		_ = s.repo.UpdateStream(ctx, stream) // best-effort
	}
	count, _ := s.streamCache.GetViewerCount(ctx, username)
	stream.ViewerCount = count
	info := stream.ToPublic(s.hlsBaseURL)
	return &info, nil
}

func (s *streamService) GetLiveStreams(ctx context.Context) ([]*domain.StreamPublicInfo, error) {
	streams, err := s.repo.GetLiveStreams(ctx)
	if err != nil {
		return nil, fmt.Errorf("get live streams: %w", err)
	}
	var result []*domain.StreamPublicInfo
	for _, stream := range streams {
		// Skip/purge stale live entries
		if stream.Status == domain.StreamStatusLive && !s.transcoder.IsRunning(stream.Username) {
			now := time.Now()
			stream.Status = domain.StreamStatusEnded
			stream.EndedAt = &now
			_ = s.streamCache.SetOffline(ctx, stream.Username)
			_ = s.repo.UpdateStream(ctx, stream) // best-effort
			continue
		}
		count, _ := s.streamCache.GetViewerCount(ctx, stream.Username)
		stream.ViewerCount = count
		info := stream.ToPublic(s.hlsBaseURL)
		result = append(result, &info)
	}
	return result, nil
}

func (s *streamService) UpdateStreamSettings(ctx context.Context, userID uuid.UUID, input domain.UpdateStreamInput) error {
	stream, err := s.repo.GetStreamByUserID(ctx, userID)
	if err != nil {
		return ErrStreamNotFound
	}
	stream.Title = input.Title
	stream.Category = input.Category
	return s.repo.UpdateStream(ctx, stream)
}

func (s *streamService) HandleStreamStart(ctx context.Context, streamKey string) (io.WriteCloser, error) {
	key, err := s.repo.GetStreamKeyByKey(ctx, streamKey)
	if err != nil {
		return nil, ErrInvalidStreamKey
	}
	stream, err := s.repo.GetStreamByUserID(ctx, key.UserID)
	if err != nil {
		return nil, fmt.Errorf("get stream: %w", err)
	}
	// Prevent duplicate transcoding when DB is already live and the transcoder is running.
	if stream.Status == domain.StreamStatusLive && s.transcoder.IsRunning(stream.Username) {
		return nil, ErrAlreadyLive
	}

	// Start ffmpeg first. Only mark the stream as live in DB/cache after the transcoder is actually running.
	pipe, err := s.transcoder.Start(stream.Username)
	if err != nil {
		return nil, fmt.Errorf("start transcoder: %w", err)
	}

	now := time.Now()
	stream.Status = domain.StreamStatusLive
	stream.StartedAt = &now
	stream.EndedAt = nil
	if err := s.repo.UpdateStream(ctx, stream); err != nil {
		// Best-effort cleanup so we don't keep a running ffmpeg without DB "live".
		s.transcoder.Stop(stream.Username)
		return nil, fmt.Errorf("update stream: %w", err)
	}
	if err := s.streamCache.SetLive(ctx, stream.Username, stream.ID.String()); err != nil {
		s.transcoder.Stop(stream.Username)
		return nil, fmt.Errorf("set live cache: %w", err)
	}

	if s.pub != nil {
		_ = s.pub.PublishStreamStarted(ctx, domain.StreamStartedEvent{
			EventType: "stream.started",
			StreamID:  stream.ID, UserID: stream.UserID,
			Username: stream.Username, Title: stream.Title, StartedAt: now,
		})
	}

	if s.pub != nil {
		s.pub.PublishStreamEnded(ctx, domain.StreamEndedEvent{
			EventType: "stream.ended",
			StreamID:  stream.ID, UserID: stream.UserID,
			Username: stream.Username, EndedAt: now,
		})
	}
	return pipe, nil
}

func (s *streamService) HandleStreamEnd(ctx context.Context, streamKey string) {
	key, err := s.repo.GetStreamKeyByKey(ctx, streamKey)
	if err != nil {
		return
	}
	stream, err := s.repo.GetStreamByUserID(ctx, key.UserID)
	if err != nil {
		return
	}
	now := time.Now()
	stream.Status = domain.StreamStatusEnded
	stream.EndedAt = &now
	s.repo.UpdateStream(ctx, stream)
	s.streamCache.SetOffline(ctx, stream.Username)
	s.transcoder.Stop(stream.Username)
	if s.pub != nil {
		s.pub.PublishStreamEnded(ctx, domain.StreamEndedEvent{
			StreamID: stream.ID, UserID: stream.UserID,
			Username: stream.Username, EndedAt: now,
		})
	}
}

func (s *streamService) JoinStream(ctx context.Context, username string) (int, error) {
	return s.streamCache.IncrementViewers(ctx, username)
}

func (s *streamService) LeaveStream(ctx context.Context, username string) (int, error) {
	return s.streamCache.DecrementViewers(ctx, username)
}

func newStreamKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "sk_live_" + strings.ToLower(hex.EncodeToString(b))
}
