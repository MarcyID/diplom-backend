package repository

import (
	"context"
	"diplomM/internal/model/auth"
	"time"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *auth.User) (*auth.User, error)
	GetByID(ctx context.Context, id int64) (*auth.User, error)
	GetByEmail(ctx context.Context, email string) (*auth.User, error)
	GetByUsername(ctx context.Context, username string) (*auth.User, error)
	Update(ctx context.Context, user *auth.User) error
	UpdateGenrePreferences(ctx context.Context, userID int64, genreIDs []int64) error
}

// SessionRepository определяет интерфейс для работы с сессиями
type SessionRepository interface {
	Create(ctx context.Context, session *auth.Session) (*auth.Session, error)
	GetByToken(ctx context.Context, tokenHash string) (*auth.Session, error)
	Delete(ctx context.Context, id int64) error
	DeleteByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context, before time.Time) error
}
