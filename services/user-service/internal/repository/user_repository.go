package repository

import (
	"context"
	"errors"
	"time"

	"github.com/VladUrsul/livestream-platform/services/user-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("profile not found")
)

type UserRepository interface {
	CreateProfile(ctx context.Context, p *domain.Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	GetByUsername(ctx context.Context, username string) (*domain.Profile, error)
	Update(ctx context.Context, p *domain.Profile) error
	Search(ctx context.Context, query string, limit int) ([]*domain.SearchResult, error)
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)
	GetFollowing(ctx context.Context, userID uuid.UUID) ([]*domain.SearchResult, error)
	SetLiveStatus(ctx context.Context, userID uuid.UUID, isLive bool) error
	GetFollowerIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type postgresRepo struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) UserRepository { return &postgresRepo{db: db} }

func (r *postgresRepo) CreateProfile(ctx context.Context, p *domain.Profile) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		INSERT INTO profiles (user_id,username,email,display_name,bio,avatar_url,followers,following,is_live,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,0,0,false,$7,$8)
		ON CONFLICT (user_id) DO NOTHING`,
		p.UserID, p.Username, p.Email, p.DisplayName, p.Bio, p.AvatarURL, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

// GetFollowerIDs returns a list of user IDs that follow the given user ID.
func (r *postgresRepo) GetFollowerIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx,
		`SELECT follower_id FROM follows WHERE followee_id = $1`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if ids == nil {
		return []uuid.UUID{}, nil
	}
	return ids, nil
}

func (r *postgresRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	return r.scan(ctx,
		`SELECT user_id,username,email,display_name,bio,avatar_url,followers,following,is_live,created_at,updated_at
		 FROM profiles WHERE user_id=$1`, userID)
}

func (r *postgresRepo) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	return r.scan(ctx,
		`SELECT user_id,username,email,display_name,bio,avatar_url,followers,following,is_live,created_at,updated_at
		 FROM profiles WHERE username=$1`, username)
}

func (r *postgresRepo) Update(ctx context.Context, p *domain.Profile) error {
	p.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE profiles SET display_name=$1,bio=$2,avatar_url=$3,updated_at=$4 WHERE user_id=$5`,
		p.DisplayName, p.Bio, p.AvatarURL, p.UpdatedAt, p.UserID,
	)
	return err
}

func (r *postgresRepo) Search(ctx context.Context, query string, limit int) ([]*domain.SearchResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	rows, err := r.db.Query(ctx, `
		SELECT user_id,username,display_name,avatar_url,followers,is_live
		FROM profiles
		WHERE username ILIKE $1 OR display_name ILIKE $1
		ORDER BY is_live DESC, followers DESC
		LIMIT $2`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.SearchResult
	for rows.Next() {
		s := &domain.SearchResult{}
		if err := rows.Scan(&s.UserID, &s.Username, &s.DisplayName, &s.AvatarURL, &s.Followers, &s.IsLive); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *postgresRepo) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tx.Exec(ctx, `INSERT INTO follows (follower_id,followee_id,created_at) VALUES ($1,$2,NOW()) ON CONFLICT DO NOTHING`, followerID, followeeID)
	tx.Exec(ctx, `UPDATE profiles SET followers=followers+1 WHERE user_id=$1`, followeeID)
	tx.Exec(ctx, `UPDATE profiles SET following=following+1 WHERE user_id=$1`, followerID)
	return tx.Commit(ctx)
}

func (r *postgresRepo) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	res, _ := tx.Exec(ctx, `DELETE FROM follows WHERE follower_id=$1 AND followee_id=$2`, followerID, followeeID)
	if res.RowsAffected() > 0 {
		tx.Exec(ctx, `UPDATE profiles SET followers=GREATEST(followers-1,0) WHERE user_id=$1`, followeeID)
		tx.Exec(ctx, `UPDATE profiles SET following=GREATEST(following-1,0) WHERE user_id=$1`, followerID)
	}
	return tx.Commit(ctx)
}

func (r *postgresRepo) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id=$1 AND followee_id=$2)`,
		followerID, followeeID,
	).Scan(&exists)
	return exists, err
}

func (r *postgresRepo) GetFollowing(ctx context.Context, userID uuid.UUID) ([]*domain.SearchResult, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.user_id, p.username, p.display_name, p.avatar_url, p.followers, p.is_live
		FROM follows f
		JOIN profiles p ON p.user_id = f.followee_id
		WHERE f.follower_id = $1
		ORDER BY p.is_live DESC, p.followers DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.SearchResult
	for rows.Next() {
		s := &domain.SearchResult{}
		if err := rows.Scan(&s.UserID, &s.Username, &s.DisplayName, &s.AvatarURL, &s.Followers, &s.IsLive); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *postgresRepo) SetLiveStatus(ctx context.Context, userID uuid.UUID, isLive bool) error {
	_, err := r.db.Exec(ctx, `UPDATE profiles SET is_live=$1 WHERE user_id=$2`, isLive, userID)
	return err
}

func (r *postgresRepo) scan(ctx context.Context, q string, args ...any) (*domain.Profile, error) {
	p := &domain.Profile{}
	err := r.db.QueryRow(ctx, q, args...).Scan(
		&p.UserID, &p.Username, &p.Email, &p.DisplayName,
		&p.Bio, &p.AvatarURL, &p.Followers, &p.Following,
		&p.IsLive, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}
