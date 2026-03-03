package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrKeyNotFound = errors.New("stream key not found")
)

type StreamRepository interface {
	CreateStream(ctx context.Context, stream *domain.Stream) error
	GetStreamByUserID(ctx context.Context, userID uuid.UUID) (*domain.Stream, error)
	GetStreamByUsername(ctx context.Context, username string) (*domain.Stream, error)
	GetStreamByID(ctx context.Context, id uuid.UUID) (*domain.Stream, error)
	UpdateStream(ctx context.Context, stream *domain.Stream) error
	GetLiveStreams(ctx context.Context) ([]*domain.Stream, error)
	CreateStreamKey(ctx context.Context, key *domain.StreamKey) error
	GetStreamKeyByUserID(ctx context.Context, userID uuid.UUID) (*domain.StreamKey, error)
	GetStreamKeyByKey(ctx context.Context, key string) (*domain.StreamKey, error)
	RotateStreamKey(ctx context.Context, userID uuid.UUID) (*domain.StreamKey, error)
}

type postgresStreamRepository struct {
	db *pgxpool.Pool
}

func NewPostgresStreamRepository(db *pgxpool.Pool) StreamRepository {
	return &postgresStreamRepository{db: db}
}

func (r *postgresStreamRepository) CreateStream(ctx context.Context, stream *domain.Stream) error {
	stream.ID = uuid.New()
	stream.CreatedAt = time.Now()
	stream.UpdatedAt = time.Now()
	stream.Status = domain.StreamStatusOffline
	_, err := r.db.Exec(ctx, `
		INSERT INTO streams (id, user_id, username, title, category, stream_key, status, viewer_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, stream.ID, stream.UserID, stream.Username, stream.Title, stream.Category,
		stream.StreamKey, stream.Status, 0, stream.CreatedAt, stream.UpdatedAt)
	return err
}

func (r *postgresStreamRepository) GetStreamByUserID(ctx context.Context, userID uuid.UUID) (*domain.Stream, error) {
	return r.scanOne(ctx, `
		SELECT id,user_id,username,title,category,stream_key,status,viewer_count,started_at,ended_at,created_at,updated_at
		FROM streams WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1`, userID)
}

func (r *postgresStreamRepository) GetStreamByUsername(ctx context.Context, username string) (*domain.Stream, error) {
	return r.scanOne(ctx, `
		SELECT id,user_id,username,title,category,stream_key,status,viewer_count,started_at,ended_at,created_at,updated_at
		FROM streams WHERE username=$1 ORDER BY created_at DESC LIMIT 1`, username)
}

func (r *postgresStreamRepository) GetStreamByID(ctx context.Context, id uuid.UUID) (*domain.Stream, error) {
	return r.scanOne(ctx, `
		SELECT id,user_id,username,title,category,stream_key,status,viewer_count,started_at,ended_at,created_at,updated_at
		FROM streams WHERE id=$1`, id)
}

func (r *postgresStreamRepository) UpdateStream(ctx context.Context, stream *domain.Stream) error {
	stream.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE streams
		SET title=$1,category=$2,status=$3,viewer_count=$4,started_at=$5,ended_at=$6,updated_at=$7
		WHERE id=$8
	`, stream.Title, stream.Category, stream.Status, stream.ViewerCount,
		stream.StartedAt, stream.EndedAt, stream.UpdatedAt, stream.ID)
	return err
}

func (r *postgresStreamRepository) GetLiveStreams(ctx context.Context) ([]*domain.Stream, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id,user_id,username,title,category,stream_key,status,viewer_count,started_at,ended_at,created_at,updated_at
		FROM streams WHERE status='live' ORDER BY viewer_count DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var streams []*domain.Stream
	for rows.Next() {
		s := &domain.Stream{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Username, &s.Title, &s.Category,
			&s.StreamKey, &s.Status, &s.ViewerCount, &s.StartedAt, &s.EndedAt,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		streams = append(streams, s)
	}
	return streams, nil
}

func (r *postgresStreamRepository) CreateStreamKey(ctx context.Context, key *domain.StreamKey) error {
	key.ID = uuid.New()
	key.CreatedAt = time.Now()
	key.Active = true
	_, err := r.db.Exec(ctx,
		`INSERT INTO stream_keys (id,user_id,key,active,created_at) VALUES ($1,$2,$3,$4,$5)`,
		key.ID, key.UserID, key.Key, key.Active, key.CreatedAt)
	return err
}

func (r *postgresStreamRepository) GetStreamKeyByUserID(ctx context.Context, userID uuid.UUID) (*domain.StreamKey, error) {
	k := &domain.StreamKey{}
	err := r.db.QueryRow(ctx,
		`SELECT id,user_id,key,active,created_at FROM stream_keys WHERE user_id=$1 AND active=true`, userID,
	).Scan(&k.ID, &k.UserID, &k.Key, &k.Active, &k.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrKeyNotFound
	}
	return k, err
}

func (r *postgresStreamRepository) GetStreamKeyByKey(ctx context.Context, key string) (*domain.StreamKey, error) {
	k := &domain.StreamKey{}
	err := r.db.QueryRow(ctx,
		`SELECT id,user_id,key,active,created_at FROM stream_keys WHERE key=$1 AND active=true`, key,
	).Scan(&k.ID, &k.UserID, &k.Key, &k.Active, &k.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrKeyNotFound
	}
	return k, err
}

func (r *postgresStreamRepository) RotateStreamKey(ctx context.Context, userID uuid.UUID) (*domain.StreamKey, error) {
	if _, err := r.db.Exec(ctx, `UPDATE stream_keys SET active=false WHERE user_id=$1`, userID); err != nil {
		return nil, fmt.Errorf("deactivate old key: %w", err)
	}
	newKey := &domain.StreamKey{UserID: userID, Key: generateStreamKey()}
	if err := r.CreateStreamKey(ctx, newKey); err != nil {
		return nil, fmt.Errorf("create new key: %w", err)
	}
	return newKey, nil
}

func (r *postgresStreamRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Stream, error) {
	s := &domain.Stream{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&s.ID, &s.UserID, &s.Username, &s.Title, &s.Category,
		&s.StreamKey, &s.Status, &s.ViewerCount,
		&s.StartedAt, &s.EndedAt, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan stream: %w", err)
	}
	return s, nil
}

func generateStreamKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "sk_live_" + strings.ToLower(hex.EncodeToString(b))
}
