package repository

import (
	"context"
	"diplomM/internal/database"
	"diplomM/internal/model/auth"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// sessionRepository реализует SessionRepository
type sessionRepository struct {
	db *database.PostgreSQL
}

// NewSessionRepository создает новый sessionRepository
func NewSessionRepository(db *database.PostgreSQL) SessionRepository {
	return &sessionRepository{db: db}
}

// Create создает новую сессию
func (r *sessionRepository) Create(ctx context.Context, session *auth.Session) (*auth.Session, error) {
	query := `
		INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	now := time.Now()
	err := r.db.Pool.QueryRow(
		ctx, query,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		now,
	).Scan(&session.ID, &session.CreatedAt)

	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetByToken получает сессию по хешу токена
func (r *sessionRepository) GetByToken(ctx context.Context, tokenHash string) (*auth.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, is_revoked
		FROM sessions
		WHERE refresh_token_hash = $1 AND is_revoked = FALSE AND expires_at > NOW()
	`

	session := &auth.Session{}
	err := r.db.Pool.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.IsRevoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return session, nil
}

// Delete удаляет сессию по ID
func (r *sessionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// DeleteByUserID удаляет все сессии пользователя
func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	return err
}

// DeleteExpired удаляет истекшие сессии
func (r *sessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	_, err := r.db.Pool.Exec(ctx, query, before)
	return err
}
