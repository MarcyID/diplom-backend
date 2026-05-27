package collection

import (
	"diplomM/internal/model/kinopoisk"
	"time"
)

// Collection представляет подборку фильмов
type Collection struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CollectionFilm представляет фильм в подборке
type CollectionFilm struct {
	ID           int64     `json:"id"`
	CollectionID int64     `json:"collection_id"`
	FilmID       int64     `json:"film_id"` // kinopoiskId
	Position     int       `json:"position"`
	AddedAt      time.Time `json:"added_at"`
}

// CollectionWithFilms подборка с полной информацией о фильмах
type CollectionWithFilms struct {
	ID          int64                 `json:"id"`
	UserID      int64                 `json:"user_id"`
	Title       string                `json:"title"`
	Description *string               `json:"description,omitempty"`
	IsPublic    bool                  `json:"is_public"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Films       []kinopoisk.FilmBasic `json:"films"`
}

// CreateCollectionRequest запрос на создание подборки
type CreateCollectionRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty" binding:"max=1000"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateCollectionRequest запрос на обновление подборки
type UpdateCollectionRequest struct {
	Title       string `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
	Description string `json:"description,omitempty" binding:"max=1000"`
	IsPublic    *bool  `json:"is_public,omitempty"`
}

// AddFilmToCollectionRequest запрос на добавление фильма в подборку
type AddFilmToCollectionRequest struct {
	FilmID   int64 `json:"film_id" binding:"required"`
	Position *int  `json:"position,omitempty"` // опционально, если не указано - добавляется в конец
}

// CollectionInfo краткая информация о подборке (для списков)
type CollectionInfo struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	IsPublic    bool      `json:"is_public"`
	FilmsCount  int       `json:"films_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
