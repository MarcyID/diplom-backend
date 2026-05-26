package kinopoisk

// KinopoiskStaffResponse - ответ со списком актёров/режиссёров фильма
// GET /api/v1/staff?filmId={id}
type KinopoiskStaffResponse struct {
	StaffID        int64   `json:"staffId"`
	NameRU         *string `json:"nameRu"`
	NameEN         *string `json:"nameEn"`
	Description    *string `json:"description"`
	PosterURL      string  `json:"posterUrl"`
	ProfessionText string  `json:"professionText"`
	ProfessionKey  string  `json:"professionKey"`
}

// KinopoiskPersonResponse - ответ с данными о персоне
// GET /api/v1/staff/{id}
type KinopoiskPersonResponse struct {
	PersonID   int64        `json:"personId"`
	WebURL     *string      `json:"webUrl"`
	NameRU     *string      `json:"nameRu"`
	NameEN     *string      `json:"nameEn"`
	Sex        *string      `json:"sex"`
	PosterURL  string       `json:"posterUrl"`
	Growth     *int         `json:"growth"`
	Birthday   *string      `json:"birthday"`
	Death      *string      `json:"death"`
	Birthplace *string      `json:"birthplace"`
	Deathplace *string      `json:"deathplace"`
	Profession *string      `json:"profession"`
	Facts      []string     `json:"facts"`
	Films      []PersonFilm `json:"films"`
	HasAwards  *int         `json:"hasAwards"`
	Spouses    []Spouse     `json:"spouses"`
}

// PersonFilm - фильм с участием персоны
type PersonFilm struct {
	FilmID         int64   `json:"filmId"`
	NameRU         *string `json:"nameRu"`
	NameEN         *string `json:"nameEn"`
	PosterURL      *string `json:"posterUrl"`
	ProfessionKey  string  `json:"professionKey"`
	ProfessionText *string `json:"professionText"`
	ReleaseYear    *int    `json:"releaseYear"`
	Description    *string `json:"description"`
	General        *bool   `json:"general"`
	Rating         *string `json:"rating"`
}

// Spouse - супруг/супруга персоны
type Spouse struct {
	PersonID       int64   `json:"personId"`
	Name           *string `json:"name"`
	Divorced       *bool   `json:"divorced"`
	DivorcedReason *string `json:"divorcedReason"`
	Sex            *string `json:"sex"`
	Children       *int    `json:"children"`
	WebURL         *string `json:"webUrl"`
	Relation       *string `json:"relation"`
}

// PersonByNameItem - элемент поиска персоны по имени
type PersonByNameItem struct {
	KinopoiskID int64   `json:"kinopoiskId"`
	WebURL      string  `json:"webUrl"`
	NameRU      *string `json:"nameRu"`
	NameEN      *string `json:"nameEn"`
	Sex         *string `json:"sex"`
	PosterURL   string  `json:"posterUrl"`
}

// KinopoiskPersonByNameResponse - ответ на поиск персоны по имени
// GET /api/v1/persons
type KinopoiskPersonByNameResponse struct {
	Total int                `json:"total"`
	Items []PersonByNameItem `json:"items"`
}
