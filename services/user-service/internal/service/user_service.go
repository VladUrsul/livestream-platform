package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/VladUrsul/livestream-platform/services/user-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("user not found")
	ErrCannotFollow = errors.New("cannot follow yourself")
)

type UserService interface {
	GetProfile(ctx context.Context, username string) (*domain.Profile, error)
	GetProfileByID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input domain.UpdateProfileInput) (*domain.Profile, error)
	Search(ctx context.Context, query string) ([]*domain.SearchResult, error)
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)
	GetFollowing(ctx context.Context, userID uuid.UUID) ([]*domain.SearchResult, error)
	CreateFromEvent(ctx context.Context, e domain.UserRegisteredEvent) error
	SetLiveStatus(ctx context.Context, userID uuid.UUID, isLive bool) error
}

type svc struct{ repo repository.UserRepository }

func New(repo repository.UserRepository) UserService { return &svc{repo} }

func (s *svc) GetProfile(ctx context.Context, username string) (*domain.Profile, error) {
	p, err := s.repo.GetByUsername(ctx, username)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *svc) GetProfileByID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *svc) UpdateProfile(ctx context.Context, userID uuid.UUID, input domain.UpdateProfileInput) (*domain.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	if input.DisplayName != "" {
		p.DisplayName = input.DisplayName
	}
	p.Bio = input.Bio
	if input.AvatarURL != "" {
		p.AvatarURL = input.AvatarURL
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}
	return p, nil
}

func (s *svc) Search(ctx context.Context, query string) ([]*domain.SearchResult, error) {
	if len(query) < 1 {
		return []*domain.SearchResult{}, nil
	}
	results, err := s.repo.Search(ctx, query, 10)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []*domain.SearchResult{}, nil
	}
	return results, nil
}

func (s *svc) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == followeeID {
		return ErrCannotFollow
	}
	return s.repo.Follow(ctx, followerID, followeeID)
}

func (s *svc) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	return s.repo.Unfollow(ctx, followerID, followeeID)
}

func (s *svc) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
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

func (s *svc) CreateFromEvent(ctx context.Context, e domain.UserRegisteredEvent) error {
	return s.repo.CreateProfile(ctx, &domain.Profile{
		UserID:      e.UserID,
		Username:    e.Username,
		Email:       e.Email,
		DisplayName: e.Username,
		CreatedAt:   e.CreatedAt,
	})
}

func (s *svc) SetLiveStatus(ctx context.Context, userID uuid.UUID, isLive bool) error {
	return s.repo.SetLiveStatus(ctx, userID, isLive)
}
