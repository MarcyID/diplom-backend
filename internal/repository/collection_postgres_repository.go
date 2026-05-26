package repository

import (
	"context"
	"diplomM/internal/database"
	"diplomM/internal/model/collection"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// postgresCollectionRepository реализует CollectionRepository для PostgreSQL
type postgresCollectionRepository struct {
	db *database.PostgreSQL
}

// NewPostgresCollectionRepository создает новый repository для работы с подборками
func NewPostgresCollectionRepository(db *database.PostgreSQL) CollectionRepository {
	return &postgresCollectionRepository{db: db}
}

// Create создает новую подборку
func (r *postgresCollectionRepository) Create(ctx context.Context, coll *collection.Collection) (*collection.Collection, error) {
	query := `
		INSERT INTO collections (user_id, title, description, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(ctx, query,
		coll.UserID,
		coll.Title,
		coll.Description,
		coll.IsPublic,
	).Scan(&coll.ID, &coll.CreatedAt, &coll.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return coll, nil
}

// GetByID получает подборку по ID
func (r *postgresCollectionRepository) GetByID(ctx context.Context, id int64) (*collection.Collection, error) {
	query := `
		SELECT id, user_id, title, description, is_public, created_at, updated_at
		FROM collections
		WHERE id = $1
	`

	coll := &collection.Collection{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&coll.ID,
		&coll.UserID,
		&coll.Title,
		&coll.Description,
		&coll.IsPublic,
		&coll.CreatedAt,
		&coll.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	return coll, nil
}

// GetByUserID получает все подборки пользователя
func (r *postgresCollectionRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*collection.CollectionInfo, int, error) {
	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM collections WHERE user_id = $1`
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count collections: %w", err)
	}

	query := `
		SELECT c.id, c.user_id, c.title, c.description, c.is_public, c.created_at, c.updated_at,
		       COUNT(cf.film_id) as films_count
		FROM collections c
		LEFT JOIN collection_films cf ON c.id = cf.collection_id
		WHERE c.user_id = $1
		GROUP BY c.id
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get collections: %w", err)
	}
	defer rows.Close()

	collections := make([]*collection.CollectionInfo, 0)
	for rows.Next() {
		coll := &collection.CollectionInfo{}
		err := rows.Scan(
			&coll.ID,
			&coll.UserID,
			&coll.Title,
			&coll.Description,
			&coll.IsPublic,
			&coll.CreatedAt,
			&coll.UpdatedAt,
			&coll.FilmsCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan collection: %w", err)
		}
		collections = append(collections, coll)
	}

	return collections, total, nil
}

// GetPublicByUserID получает публичные подборки пользователя
func (r *postgresCollectionRepository) GetPublicByUserID(ctx context.Context, userID int64, limit, offset int) ([]*collection.CollectionInfo, int, error) {
	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM collections WHERE user_id = $1 AND is_public = true`
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count collections: %w", err)
	}

	query := `
		SELECT c.id, c.user_id, c.title, c.description, c.is_public, c.created_at, c.updated_at,
		       COUNT(cf.film_id) as films_count
		FROM collections c
		LEFT JOIN collection_films cf ON c.id = cf.collection_id
		WHERE c.user_id = $1 AND c.is_public = true
		GROUP BY c.id
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get collections: %w", err)
	}
	defer rows.Close()

	collections := make([]*collection.CollectionInfo, 0)
	for rows.Next() {
		coll := &collection.CollectionInfo{}
		err := rows.Scan(
			&coll.ID,
			&coll.UserID,
			&coll.Title,
			&coll.Description,
			&coll.IsPublic,
			&coll.CreatedAt,
			&coll.UpdatedAt,
			&coll.FilmsCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan collection: %w", err)
		}
		collections = append(collections, coll)
	}

	return collections, total, nil
}

// Update обновляет подборку
func (r *postgresCollectionRepository) Update(ctx context.Context, coll *collection.Collection) error {
	query := `
		UPDATE collections
		SET title = $1, description = $2, is_public = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := r.db.Pool.Exec(ctx, query,
		coll.Title,
		coll.Description,
		coll.IsPublic,
		coll.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("collection not found")
	}

	return nil
}

// Delete удаляет подборку
func (r *postgresCollectionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM collections WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("collection not found")
	}

	return nil
}

// AddFilm добавляет фильм в подборку
func (r *postgresCollectionRepository) AddFilm(ctx context.Context, film *collection.CollectionFilm) error {
	// Если позиция не указана, добавляем в конец
	if film.Position == 0 {
		maxPosQuery := `SELECT COALESCE(MAX(position), -1) FROM collection_films WHERE collection_id = $1`
		err := r.db.Pool.QueryRow(ctx, maxPosQuery, film.CollectionID).Scan(&film.Position)
		if err != nil {
			return fmt.Errorf("failed to get max position: %w", err)
		}
		film.Position++
	}

	query := `
		INSERT INTO collection_films (collection_id, film_id, position, added_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (collection_id, film_id) DO NOTHING
	`

	result, err := r.db.Pool.Exec(ctx, query,
		film.CollectionID,
		film.FilmID,
		film.Position,
	)
	if err != nil {
		return fmt.Errorf("failed to add film to collection: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("film already in collection")
	}

	return nil
}

// RemoveFilm удаляет фильм из подборки
func (r *postgresCollectionRepository) RemoveFilm(ctx context.Context, collectionID, filmID int64) error {
	query := `DELETE FROM collection_films WHERE collection_id = $1 AND film_id = $2`

	result, err := r.db.Pool.Exec(ctx, query, collectionID, filmID)
	if err != nil {
		return fmt.Errorf("failed to remove film from collection: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("film not found in collection")
	}

	return nil
}

// GetFilms получает все фильмы из подборки
func (r *postgresCollectionRepository) GetFilms(ctx context.Context, collectionID int64) ([]*collection.CollectionFilm, error) {
	query := `
		SELECT id, collection_id, film_id, position, added_at
		FROM collection_films
		WHERE collection_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get films: %w", err)
	}
	defer rows.Close()

	films := make([]*collection.CollectionFilm, 0)
	for rows.Next() {
		film := &collection.CollectionFilm{}
		err := rows.Scan(
			&film.ID,
			&film.CollectionID,
			&film.FilmID,
			&film.Position,
			&film.AddedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan film: %w", err)
		}
		films = append(films, film)
	}

	return films, nil
}

// ReorderFilms изменяет порядок фильмов в подборке
func (r *postgresCollectionRepository) ReorderFilms(ctx context.Context, collectionID int64, filmPositions map[int64]int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE collection_films
		SET position = $1
		WHERE collection_id = $2 AND film_id = $3
	`

	for filmID, position := range filmPositions {
		_, err := tx.Exec(ctx, query, position, collectionID, filmID)
		if err != nil {
			return fmt.Errorf("failed to update position: %w", err)
		}
	}

	return tx.Commit(ctx)
}
