package service

import (
	"context"
	"errors"
	"log"

	"github.com/VladUrsul/livestream-platform/services/user-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("user not found")
	ErrCannotFollow = errors.New("cannot follow yourself")
)

type EventPublisher interface {
	Publish(routingKey string, payload any) error
}

type UserService interface {
	GetProfile(ctx context.Context, username string) (*domain.Profile, error)
	GetProfileByID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input domain.UpdateProfileInput) (*domain.Profile, error)
	Search(ctx context.Context, query string) ([]*domain.SearchResult, error)
	Follow(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) (bool, error)
	GetFollowing(ctx context.Context, userID uuid.UUID) ([]*domain.SearchResult, error)
	CreateFromEvent(ctx context.Context, evt domain.UserRegisteredEvent) error
	SetLiveStatus(ctx context.Context, userID uuid.UUID, isLive bool) error
	GetFollowerIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type svc struct {
	repo repository.UserRepository
	pub  EventPublisher
}

func NewUserService(repo repository.UserRepository, pub EventPublisher) UserService {
	return &svc{repo: repo, pub: pub}
}

func (s *svc) GetFollowerIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetFollowerIDs(ctx, userID)
}

// keep old constructor name working too
func New(repo repository.UserRepository) UserService {
	return &svc{repo: repo, pub: nil}
}

func (s *svc) GetProfile(ctx context.Context, username string) (*domain.Profile, error) {
	p, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *svc) GetProfileByID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *svc) UpdateProfile(ctx context.Context, userID uuid.UUID, input domain.UpdateProfileInput) (*domain.Profile, error) {
	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	if input.DisplayName != "" {
		existing.DisplayName = input.DisplayName
	}
	if input.Bio != "" {
		existing.Bio = input.Bio
	}
	if input.AvatarURL != "" {
		existing.AvatarURL = input.AvatarURL
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *svc) Search(ctx context.Context, query string) ([]*domain.SearchResult, error) {
	if len(query) < 1 {
		return []*domain.SearchResult{}, nil
	}
	return s.repo.Search(ctx, query, 10)
}

func (s *svc) Follow(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) error {
	if followerID == followeeID {
		return ErrCannotFollow
	}
	if err := s.repo.Follow(ctx, followerID, followeeID); err != nil {
		return err
	}
	// Publish user.followed event
	if s.pub != nil {
		follower, _ := s.repo.GetByUserID(ctx, followerID)
		followerUsername := ""
		if follower != nil {
			followerUsername = follower.Username
		}
		if err := s.pub.Publish("user.followed", map[string]any{
			"follower_id":       followerID,
			"follower_username": followerUsername,
			"followee_id":       followeeID,
		}); err != nil {
			log.Printf("[UserService] failed to publish user.followed: %v", err)
		}
	}
	return nil
}

func (s *svc) Unfollow(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) error {
	return s.repo.Unfollow(ctx, followerID, followeeID)
}

func (s *svc) IsFollowing(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) (bool, error) {
	return s.repo.IsFollowing(ctx, followerID, followeeID)
}

func (s *svc) GetFollowing(ctx context.Context, userID uuid.UUID) ([]*domain.SearchResult, error) {
	results, err := s.repo.GetFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []*domain.SearchResult{}, nil
	}
	return results, nil
}

func (s *svc) CreateFromEvent(ctx context.Context, evt domain.UserRegisteredEvent) error {
	return s.repo.CreateProfile(ctx, &domain.Profile{
		UserID:      evt.UserID,
		Username:    evt.Username,
		Email:       evt.Email,
		DisplayName: evt.Username,
	})
}

func (s *svc) SetLiveStatus(ctx context.Context, userID uuid.UUID, isLive bool) error {
	return s.repo.SetLiveStatus(ctx, userID, isLive)
}
