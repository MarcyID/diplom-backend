package repository

import (
	"context"
	"diplomM/internal/database"
	"diplomM/internal/model/auth"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// userRepository реализует UserRepository
type userRepository struct {
	db *database.PostgreSQL
}

// NewUserRepository создает новый userRepository
func NewUserRepository(db *database.PostgreSQL) UserRepository {
	return &userRepository{db: db}
}

// Create создает нового пользователя
func (r *userRepository) Create(ctx context.Context, user *auth.User) (*auth.User, error) {
	query := `
		INSERT INTO users (email, username, password_hash, full_name, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.Pool.QueryRow(
		ctx, query,
		user.Email,
		user.Username,
		user.Password,
		user.FullName,
		user.AvatarURL,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID получает пользователя по ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*auth.User, error) {
	query := `
		SELECT id, email, username, password_hash, full_name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &auth.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.FullName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail получает пользователя по email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, email, username, password_hash, full_name, avatar_url, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &auth.User{}
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.FullName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetByUsername получает пользователя по username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `
		SELECT id, email, username, password_hash, full_name, avatar_url, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &auth.User{}
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.FullName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// Update обновляет данные пользователя
func (r *userRepository) Update(ctx context.Context, user *auth.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, password_hash = $3, full_name = $4, avatar_url = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Pool.Exec(ctx, query,
		user.Email,
		user.Username,
		user.Password,
		user.FullName,
		user.AvatarURL,
		time.Now(),
		user.ID,
	)

	return err
}
