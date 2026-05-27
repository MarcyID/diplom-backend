package service

import (
	"context"
	"diplomM/internal/model/favorite"
	"diplomM/internal/repository"
	"sync"
)

// FavoriteService сервис для работы с избранным
type FavoriteService struct {
	favoriteRepo repository.FavoriteRepository
	kinopoisk    *KinopoiskClient
}

// NewFavoriteService создает новый FavoriteService
func NewFavoriteService(
	favoriteRepo repository.FavoriteRepository,
	kinopoisk *KinopoiskClient,
) *FavoriteService {
	return &FavoriteService{
		favoriteRepo: favoriteRepo,
		kinopoisk:    kinopoisk,
	}
}

// AddFilm добавляет фильм в избранное
func (s *FavoriteService) AddFilm(ctx context.Context, userID int64, filmID int64) error {
	fav := &favorite.Favorite{
		UserID:     userID,
		ObjectType: favorite.FavoriteTypeFilm,
		ObjectID:   filmID,
	}

	return s.favoriteRepo.Add(ctx, fav)
}

// AddPerson добавляет персону в избранное
func (s *FavoriteService) AddPerson(ctx context.Context, userID int64, personID int64) error {
	fav := &favorite.Favorite{
		UserID:     userID,
		ObjectType: favorite.FavoriteTypePerson,
		ObjectID:   personID,
	}

	return s.favoriteRepo.Add(ctx, fav)
}

// RemoveFilm удаляет фильм из избранного
func (s *FavoriteService) RemoveFilm(ctx context.Context, userID int64, filmID int64) error {
	return s.favoriteRepo.Remove(ctx, userID, favorite.FavoriteTypeFilm, filmID)
}

// RemovePerson удаляет персону из избранного
func (s *FavoriteService) RemovePerson(ctx context.Context, userID int64, personID int64) error {
	return s.favoriteRepo.Remove(ctx, userID, favorite.FavoriteTypePerson, personID)
}

// GetFavorites получает все избранные объекты пользователя с полной информацией
func (s *FavoriteService) GetFavorites(ctx context.Context, userID int64, page, pageSize int) ([]favorite.FavoriteItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Получаем список favorites
	favorites, total, err := s.favoriteRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Загружаем данные из Kinopoisk API параллельно (батчами по 10)
	items := make([]favorite.FavoriteItem, 0, len(favorites))
	const batchSize = 10

	for i := 0; i < len(favorites); i += batchSize {
		end := i + batchSize
		if end > len(favorites) {
			end = len(favorites)
		}
		batch := favorites[i:end]

		batchItems, err := s.loadFavoritesBatch(ctx, batch)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, batchItems...)
	}

	return items, total, nil
}

// loadFavoritesBatch загружает данные для батча избранных объектов параллельно
func (s *FavoriteService) loadFavoritesBatch(ctx context.Context, favorites []*favorite.Favorite) ([]favorite.FavoriteItem, error) {
	items := make([]favorite.FavoriteItem, len(favorites))
	var wg sync.WaitGroup
	errChan := make(chan error, len(favorites))

	for i, fav := range favorites {
		wg.Add(1)
		go func(idx int, f *favorite.Favorite) {
			defer wg.Done()

			item := favorite.FavoriteItem{
				ObjectType: f.ObjectType,
				ObjectID:   f.ObjectID,
				CreatedAt:  f.CreatedAt,
			}

			switch f.ObjectType {
			case favorite.FavoriteTypeFilm:
				filmInfo, err := s.kinopoisk.GetFilmByID(f.ObjectID)
				if err != nil {
					// Если фильм не найден, пропускаем
					return
				}

				item.FilmData = &favorite.FilmFavoriteData{
					KinopoiskID:      filmInfo.KinopoiskID,
					NameRU:           filmInfo.NameRU,
					NameEN:           filmInfo.NameEN,
					PosterURL:        filmInfo.PosterURL,
					PosterURLPreview: filmInfo.PosterURLPreview,
					Year:             filmInfo.Year,
					RatingKinopoisk:  filmInfo.RatingKinopoisk,
					Type:             filmInfo.Type,
				}

			case favorite.FavoriteTypePerson:
				personInfo, err := s.kinopoisk.GetPersonByID(f.ObjectID)
				if err != nil {
					// Если персона не найдена, пропускаем
					return
				}

				// Получаем профессию из основного поля profession
				profession := ""
				if personInfo.Profession != nil {
					profession = *personInfo.Profession
				}

				item.PersonData = &favorite.PersonFavoriteData{
					PersonID:   personInfo.PersonID,
					NameRU:     personInfo.NameRU,
					NameEN:     personInfo.NameEN,
					PosterURL:  personInfo.PosterURL,
					Profession: profession,
				}
			}

			items[idx] = item
		}(i, fav)
	}

	// Ждем завершения всех горутин
	wg.Wait()
	close(errChan)

	// Проверяем ошибки
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// Фильтруем пустые элементы (где данные не загрузились)
	filteredItems := make([]favorite.FavoriteItem, 0, len(items))
	for _, item := range items {
		if item.FilmData != nil || item.PersonData != nil {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems, nil
}

// ToggleFilm добавляет или удаляет фильм из избранного
func (s *FavoriteService) ToggleFilm(ctx context.Context, userID int64, filmID int64) (bool, error) {
	exists, err := s.favoriteRepo.Exists(ctx, userID, favorite.FavoriteTypeFilm, filmID)
	if err != nil {
		return false, err
	}

	if exists {
		err = s.RemoveFilm(ctx, userID, filmID)
		if err != nil {
			return false, err
		}
		return false, nil
	} else {
		err = s.AddFilm(ctx, userID, filmID)
		if err != nil {
			// "already in favorites" игнорируем
			if err.Error() == "already in favorites" {
				return true, nil
			}
			return false, err
		}
		return true, nil
	}
}

// TogglePerson добавляет или удаляет персону из избранного
func (s *FavoriteService) TogglePerson(ctx context.Context, userID int64, personID int64) (bool, error) {
	exists, err := s.favoriteRepo.Exists(ctx, userID, favorite.FavoriteTypePerson, personID)
	if err != nil {
		return false, err
	}

	if exists {
		err = s.RemovePerson(ctx, userID, personID)
		if err != nil {
			return false, err
		}
		return false, nil
	} else {
		err = s.AddPerson(ctx, userID, personID)
		if err != nil {
			// "already in favorites" игнорируем
			if err.Error() == "already in favorites" {
				return true, nil
			}
			return false, err
		}
		return true, nil
	}
}
