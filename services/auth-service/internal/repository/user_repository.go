package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors — returned by the repository and handled by the service layer.
var (
	ErrNotFound          = errors.New("record not found")
	ErrDuplicateEmail    = errors.New("email already registered")
	ErrDuplicateUsername = errors.New("username already taken")
)

// UserRepository defines the persistence contract for users.
// Using an interface makes it easy to swap implementations or mock in tests.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

type postgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgreSQL user repository.
func NewPostgresUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return mapDBError(err)
	}
	return nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`, email,
	).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`, username,
	).Scan(&exists)
	return exists, err
}

func mapDBError(err error) error {
	msg := err.Error()
	if strings.Contains(msg, "users_email_key") {
		return ErrDuplicateEmail
	}
	if strings.Contains(msg, "users_username_key") {
		return ErrDuplicateUsername
	}
	return fmt.Errorf("database error: %w", err)
}
