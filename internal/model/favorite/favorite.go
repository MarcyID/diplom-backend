package favorite

import "time"

// FavoriteType - тип объекта в избранном
type FavoriteType string

const (
	FavoriteTypeFilm   FavoriteType = "film"
	FavoriteTypePerson FavoriteType = "person"
)

// Favorite представляет объект в избранном пользователя
type Favorite struct {
	ID         int64        `json:"id"`
	UserID     int64        `json:"user_id"`
	ObjectType FavoriteType `json:"object_type"`
	ObjectID   int64        `json:"object_id"` // kinopoiskId фильма или personId персоны
	CreatedAt  time.Time    `json:"created_at"`
}

// FavoriteItem - объект в избранном с полной информацией
type FavoriteItem struct {
	ObjectType FavoriteType `json:"object_type"`
	ObjectID   int64        `json:"object_id"`
	CreatedAt  time.Time    `json:"created_at"`
	// Film данные (если object_type = "film")
	FilmData *FilmFavoriteData `json:"film_data,omitempty"`
	// Person данные (если object_type = "person")
	PersonData *PersonFavoriteData `json:"person_data,omitempty"`
}

// FilmFavoriteData - информация о фильме в избранном
type FilmFavoriteData struct {
	KinopoiskID      int64    `json:"kinopoiskId"`
	NameRU           *string  `json:"nameRu"`
	NameEN           *string  `json:"nameEn"`
	PosterURL        string   `json:"posterUrl"`
	PosterURLPreview string   `json:"posterUrlPreview"`
	Year             *int     `json:"year"`
	RatingKinopoisk  *float64 `json:"ratingKinopoisk"`
	Type             string   `json:"type"`
}

// PersonFavoriteData - информация о персоне в избранном
type PersonFavoriteData struct {
	PersonID   int64   `json:"personId"`
	NameRU     *string `json:"nameRu"`
	NameEN     *string `json:"nameEn"`
	PosterURL  string  `json:"posterUrl"`
	Profession string  `json:"professionText"`
}

// AddFavoriteRequest - запрос на добавление в избранное
type AddFavoriteRequest struct {
	ObjectType FavoriteType `json:"object_type" binding:"required,oneof=film person"`
	ObjectID   int64        `json:"object_id" binding:"required"`
}

// FavoriteListResponse - ответ со списком избранного
type FavoriteListResponse struct {
	Total int            `json:"total"`
	Items []FavoriteItem `json:"items"`
}
