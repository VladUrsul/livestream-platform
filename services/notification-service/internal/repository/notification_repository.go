package repository

import (
	"context"
	"time"

	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type postgresRepo struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) NotificationRepository {
	return &postgresRepo{db}
}

func (r *postgresRepo) Create(ctx context.Context, n *domain.Notification) error {
	n.ID = uuid.New()
	n.CreatedAt = time.Now()
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications
			(id, user_id, type, title, body, actor_id, actor_name, read, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		n.ID, n.UserID, n.Type, n.Title, n.Body,
		n.ActorID, n.ActorName, false, n.CreatedAt,
	)
	return err
}

func (r *postgresRepo) GetForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, type, title, body, actor_id, actor_name, read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body,
			&n.ActorID, &n.ActorName, &n.Read, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *postgresRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read=false`, userID,
	).Scan(&count)
	return count, err
}

func (r *postgresRepo) MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET read=true WHERE id=$1 AND user_id=$2`, id, userID,
	)
	return err
}

func (r *postgresRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET read=true WHERE user_id=$1 AND read=false`, userID,
	)
	return err
}

func (r *postgresRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM notifications WHERE read=true AND created_at < $1`, before,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
