package service

import (
	"context"
	"diplomM/internal/model/collection"
	"diplomM/internal/model/kinopoisk"
	"diplomM/internal/repository"
	"errors"
)

// CollectionService сервис для работы с подборками
type CollectionService struct {
	collectionRepo repository.CollectionRepository
	kinopoisk      *KinopoiskClient
}

// NewCollectionService создает новый CollectionService
func NewCollectionService(
	collectionRepo repository.CollectionRepository,
	kinopoisk *KinopoiskClient,
) *CollectionService {
	return &CollectionService{
		collectionRepo: collectionRepo,
		kinopoisk:      kinopoisk,
	}
}

// CreateCollection создает новую подборку
func (s *CollectionService) CreateCollection(ctx context.Context, userID int64, req collection.CreateCollectionRequest) (*collection.Collection, error) {
	coll := &collection.Collection{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	return s.collectionRepo.Create(ctx, coll)
}

// GetCollection получает подборку по ID с полной информацией о фильмах
func (s *CollectionService) GetCollection(ctx context.Context, id int64, requestUserID *int64) (*collection.CollectionWithFilms, error) {
	// Получаем подборку
	coll, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверяем доступ: если подборка приватная, проверяем владельца
	if !coll.IsPublic {
		if requestUserID == nil || *requestUserID != coll.UserID {
			return nil, errors.New("access denied")
		}
	}

	// Получаем фильмы из подборки
	filmRecords, err := s.collectionRepo.GetFilms(ctx, id)
	if err != nil {
		return nil, err
	}

	// Получаем информацию о фильмах из Kinopoisk API
	films := make([]kinopoisk.FilmBasic, 0, len(filmRecords))
	for _, filmRecord := range filmRecords {
		filmInfo, err := s.kinopoisk.GetFilmByID(filmRecord.FilmID)
		if err != nil {
			// Если фильм не найден, пропускаем его
			continue
		}

		films = append(films, kinopoisk.FilmBasic{
			KinopoiskID:      filmInfo.KinopoiskID,
			NameRU:           filmInfo.NameRU,
			NameEN:           filmInfo.NameEN,
			NameOriginal:     filmInfo.NameOriginal,
			PosterURL:        filmInfo.PosterURL,
			PosterURLPreview: filmInfo.PosterURLPreview,
			Year:             filmInfo.Year,
			RatingKinopoisk:  filmInfo.RatingKinopoisk,
			RatingIMDB:       filmInfo.RatingIMDB,
			Type:             filmInfo.Type,
			Countries:        filmInfo.Countries,
			Genres:           filmInfo.Genres,
		})
	}

	return &collection.CollectionWithFilms{
		ID:          coll.ID,
		UserID:      coll.UserID,
		Title:       coll.Title,
		Description: coll.Description,
		IsPublic:    coll.IsPublic,
		CreatedAt:   coll.CreatedAt,
		UpdatedAt:   coll.UpdatedAt,
		Films:       films,
	}, nil
}

// GetUserCollections получает все подборки пользователя
func (s *CollectionService) GetUserCollections(ctx context.Context, userID int64, page, pageSize int) ([]*collection.CollectionInfo, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.collectionRepo.GetByUserID(ctx, userID, pageSize, offset)
}

// GetPublicUserCollections получает публичные подборки пользователя
func (s *CollectionService) GetPublicUserCollections(ctx context.Context, userID int64, page, pageSize int) ([]*collection.CollectionInfo, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.collectionRepo.GetPublicByUserID(ctx, userID, pageSize, offset)
}

// UpdateCollection обновляет подборку
func (s *CollectionService) UpdateCollection(ctx context.Context, id int64, userID int64, req collection.UpdateCollectionRequest) (*collection.Collection, error) {
	// Получаем подборку
	coll, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверяем владельца
	if coll.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Обновляем поля
	if req.Title != "" {
		coll.Title = req.Title
	}
	if req.Description != nil {
		coll.Description = req.Description
	}
	if req.IsPublic != nil {
		coll.IsPublic = *req.IsPublic
	}

	err = s.collectionRepo.Update(ctx, coll)
	if err != nil {
		return nil, err
	}

	return coll, nil
}

// DeleteCollection удаляет подборку
func (s *CollectionService) DeleteCollection(ctx context.Context, id int64, userID int64) error {
	// Получаем подборку
	coll, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Проверяем владельца
	if coll.UserID != userID {
		return errors.New("access denied")
	}

	return s.collectionRepo.Delete(ctx, id)
}

// AddFilmToCollection добавляет фильм в подборку
func (s *CollectionService) AddFilmToCollection(ctx context.Context, collectionID int64, userID int64, req collection.AddFilmToCollectionRequest) error {
	// Получаем подборку
	coll, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}

	// Проверяем владельца
	if coll.UserID != userID {
		return errors.New("access denied")
	}

	film := &collection.CollectionFilm{
		CollectionID: collectionID,
		FilmID:       req.FilmID,
	}

	if req.Position != nil {
		film.Position = *req.Position
	}

	return s.collectionRepo.AddFilm(ctx, film)
}

// RemoveFilmFromCollection удаляет фильм из подборки
func (s *CollectionService) RemoveFilmFromCollection(ctx context.Context, collectionID int64, userID int64, filmID int64) error {
	// Получаем подборку
	coll, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}

	// Проверяем владельца
	if coll.UserID != userID {
		return errors.New("access denied")
	}

	return s.collectionRepo.RemoveFilm(ctx, collectionID, filmID)
}

// ReorderCollectionFilms изменяет порядок фильмов в подборке
func (s *CollectionService) ReorderCollectionFilms(ctx context.Context, collectionID int64, userID int64, filmPositions map[int64]int) error {
	// Получаем подборку
	coll, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}

	// Проверяем владельца
	if coll.UserID != userID {
		return errors.New("access denied")
	}

	return s.collectionRepo.ReorderFilms(ctx, collectionID, filmPositions)
}

// GetCollectionFilmsCount получает количество фильмов в подборке
func (s *CollectionService) GetCollectionFilmsCount(ctx context.Context, collectionID int64) (int, error) {
	films, err := s.collectionRepo.GetFilms(ctx, collectionID)
	if err != nil {
		return 0, err
	}
	return len(films), nil
}
