package repository

import (
	"context"
	"diplomM/internal/model/collection"
)

// CollectionRepository определяет интерфейс для работы с подборками
type CollectionRepository interface {
	// Подборки
	Create(ctx context.Context, coll *collection.Collection) (*collection.Collection, error)
	GetByID(ctx context.Context, id int64) (*collection.Collection, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*collection.CollectionInfo, int, error)
	GetPublicByUserID(ctx context.Context, userID int64, limit, offset int) ([]*collection.CollectionInfo, int, error)
	Update(ctx context.Context, coll *collection.Collection) error
	Delete(ctx context.Context, id int64) error

	// Фильмы в подборке
	AddFilm(ctx context.Context, film *collection.CollectionFilm) error
	RemoveFilm(ctx context.Context, collectionID, filmID int64) error
	GetFilms(ctx context.Context, collectionID int64) ([]*collection.CollectionFilm, error)
	ReorderFilms(ctx context.Context, collectionID int64, filmPositions map[int64]int) error
}
