package repository

import (
	"context"
	"diplomM/internal/database"
	"diplomM/internal/model/favorite"
	"errors"
	"fmt"
)

// FavoriteRepository определяет интерфейс для работы с избранным
type FavoriteRepository interface {
	Add(ctx context.Context, fav *favorite.Favorite) error
	Remove(ctx context.Context, userID int64, objectType favorite.FavoriteType, objectID int64) error
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*favorite.Favorite, int, error)
	Exists(ctx context.Context, userID int64, objectType favorite.FavoriteType, objectID int64) (bool, error)
}

// postgresFavoriteRepository реализует FavoriteRepository для PostgreSQL
type postgresFavoriteRepository struct {
	db *database.PostgreSQL
}

// NewPostgresFavoriteRepository создает новый repository для работы с избранным
func NewPostgresFavoriteRepository(db *database.PostgreSQL) FavoriteRepository {
	return &postgresFavoriteRepository{db: db}
}

// Add добавляет объект в избранное
func (r *postgresFavoriteRepository) Add(ctx context.Context, fav *favorite.Favorite) error {
	query := `
		INSERT INTO favorites (user_id, object_type, object_id, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, object_type, object_id) DO NOTHING
	`

	result, err := r.db.Pool.Exec(ctx, query,
		fav.UserID,
		fav.ObjectType,
		fav.ObjectID,
	)
	if err != nil {
		return fmt.Errorf("failed to add to favorites: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("already in favorites")
	}

	return nil
}

// Remove удаляет объект из избранного
func (r *postgresFavoriteRepository) Remove(ctx context.Context, userID int64, objectType favorite.FavoriteType, objectID int64) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND object_type = $2 AND object_id = $3`

	result, err := r.db.Pool.Exec(ctx, query, userID, objectType, objectID)
	if err != nil {
		return fmt.Errorf("failed to remove from favorites: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("not in favorites")
	}

	return nil
}

// GetByUserID получает все избранные объекты пользователя
func (r *postgresFavoriteRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*favorite.Favorite, int, error) {
	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count favorites: %w", err)
	}

	query := `
		SELECT id, user_id, object_type, object_id, created_at
		FROM favorites
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get favorites: %w", err)
	}
	defer rows.Close()

	favorites := make([]*favorite.Favorite, 0)
	for rows.Next() {
		fav := &favorite.Favorite{}
		var objectType string
		err := rows.Scan(
			&fav.ID,
			&fav.UserID,
			&objectType,
			&fav.ObjectID,
			&fav.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan favorite: %w", err)
		}
		fav.ObjectType = favorite.FavoriteType(objectType)
		favorites = append(favorites, fav)
	}

	return favorites, total, nil
}

// Exists проверяет, есть ли объект в избранном
func (r *postgresFavoriteRepository) Exists(ctx context.Context, userID int64, objectType favorite.FavoriteType, objectID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND object_type = $2 AND object_id = $3)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, userID, objectType, objectID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists, nil
}
