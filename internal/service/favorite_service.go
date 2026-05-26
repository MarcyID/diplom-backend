package service

import (
	"context"
	"diplomM/internal/model/favorite"
	"diplomM/internal/repository"
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
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Получаем список favorites
	favorites, total, err := s.favoriteRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Для каждого получаем детали из Kinopoisk API
	items := make([]favorite.FavoriteItem, 0, len(favorites))
	for _, fav := range favorites {
		item := favorite.FavoriteItem{
			ObjectType: fav.ObjectType,
			ObjectID:   fav.ObjectID,
			CreatedAt:  fav.CreatedAt,
		}

		switch fav.ObjectType {
		case favorite.FavoriteTypeFilm:
			filmInfo, err := s.kinopoisk.GetFilmByID(fav.ObjectID)
			if err != nil {
				// Если фильм не найден, пропускаем
				continue
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
			personInfo, err := s.kinopoisk.GetPersonByID(fav.ObjectID)
			if err != nil {
				// Если персона не найдена, пропускаем
				continue
			}

			// Получаем профессию из первого фильма
			profession := ""
			if len(personInfo.Films) > 0 && personInfo.Films[0].ProfessionText != nil {
				profession = *personInfo.Films[0].ProfessionText
			}

			item.PersonData = &favorite.PersonFavoriteData{
				PersonID:   personInfo.PersonID,
				NameRU:     personInfo.NameRU,
				NameEN:     personInfo.NameEN,
				PosterURL:  personInfo.PosterURL,
				Profession: profession,
			}
		}

		items = append(items, item)
	}

	return items, total, nil
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
