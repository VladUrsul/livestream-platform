package repository

import (
	"context"
	"time"

	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, msg *domain.Message) error
	GetHistory(ctx context.Context, roomID string, limit int) ([]domain.Message, error)
	GetRoom(ctx context.Context, roomID string) (*domain.Room, error)
	UpsertRoom(ctx context.Context, room *domain.Room) error
	SetSlowMode(ctx context.Context, roomID string, seconds int) error
}

type postgresRepo struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) ChatRepository { return &postgresRepo{db} }

func (r *postgresRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		INSERT INTO messages (id, room_id, user_id, username, content, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		msg.ID, msg.RoomID, msg.UserID, msg.Username, msg.Content, msg.CreatedAt,
	)
	return err
}

func (r *postgresRepo) GetHistory(ctx context.Context, roomID string, limit int) ([]domain.Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, room_id, user_id, username, content, created_at
		FROM messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		roomID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Username, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	// Reverse so oldest is first
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (r *postgresRepo) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	room := &domain.Room{}
	err := r.db.QueryRow(ctx,
		`SELECT id, slow_mode, created_at FROM rooms WHERE id=$1`, roomID,
	).Scan(&room.ID, &room.SlowMode, &room.CreatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *postgresRepo) UpsertRoom(ctx context.Context, room *domain.Room) error {
	room.CreatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		INSERT INTO rooms (id, slow_mode, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`,
		room.ID, room.SlowMode, room.CreatedAt,
	)
	return err
}

func (r *postgresRepo) SetSlowMode(ctx context.Context, roomID string, seconds int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE rooms SET slow_mode=$1 WHERE id=$2`, seconds, roomID,
	)
	return err
}
