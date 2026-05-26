package kinopoisk

// FilmBasic - базовая информация о фильме (для подборок, списков)
type FilmBasic struct {
	KinopoiskID      int64     `json:"kinopoiskId"`
	NameRU           *string   `json:"nameRu"`
	NameEN           *string   `json:"nameEn"`
	NameOriginal     *string   `json:"nameOriginal"`
	PosterURL        string    `json:"posterUrl"`
	PosterURLPreview string    `json:"posterUrlPreview"`
	Year             *int      `json:"year"`
	RatingKinopoisk  *float64  `json:"ratingKinopoisk"`
	RatingIMDB       *float64  `json:"ratingImdb"`
	Type             string    `json:"type"`
	Countries        []Country `json:"countries"`
	Genres           []Genre   `json:"genres"`
}

// KinopoiskFilm - основная структура фильма из Kinopoisk API v2.2
// GET /api/v2.2/films/{id}
type KinopoiskFilm struct {
	KinopoiskID                int64          `json:"kinopoiskId"`
	KinopoiskHDID              *string        `json:"kinopoiskHDId"`
	IMDBID                     *string        `json:"imdbId"`
	NameRU                     *string        `json:"nameRu"`
	NameEN                     *string        `json:"nameEn"`
	NameOriginal               *string        `json:"nameOriginal"`
	PosterURL                  string         `json:"posterUrl"`
	PosterURLPreview           string         `json:"posterUrlPreview"`
	CoverURL                   *string        `json:"coverUrl"`
	LogoURL                    *string        `json:"logoUrl"`
	ReviewsCount               int            `json:"reviewsCount"`
	RatingGoodReview           *float64       `json:"ratingGoodReview"`
	RatingGoodReviewVoteCount  *int           `json:"ratingGoodReviewVoteCount"`
	RatingKinopoisk            *float64       `json:"ratingKinopoisk"`
	RatingKinopoiskVoteCount   *int           `json:"ratingKinopoiskVoteCount"`
	RatingIMDB                 *float64       `json:"ratingImdb"`
	RatingIMDBVoteCount        *int           `json:"ratingImdbVoteCount"`
	RatingFilmCritics          *float64       `json:"ratingFilmCritics"`
	RatingFilmCriticsVoteCount *int           `json:"ratingFilmCriticsVoteCount"`
	RatingAwait                *float64       `json:"ratingAwait"`
	RatingAwaitCount           *int           `json:"ratingAwaitCount"`
	RatingRfCritics            *float64       `json:"ratingRfCritics"`
	RatingRfCriticsVoteCount   *int           `json:"ratingRfCriticsVoteCount"`
	WebURL                     string         `json:"webUrl"`
	Year                       *int           `json:"year"`
	FilmLength                 *int           `json:"filmLength"`
	Slogan                     *string        `json:"slogan"`
	Description                string         `json:"description"`
	ShortDescription           *string        `json:"shortDescription"`
	EditorAnnotation           *string        `json:"editorAnnotation"`
	IsTicketsAvailable         bool           `json:"isTicketsAvailable"`
	ProductionStatus           *string        `json:"productionStatus"`
	Type                       string         `json:"type"`
	RatingMpaa                 *string        `json:"ratingMpaa"`
	RatingAgeLimits            *string        `json:"ratingAgeLimits"`
	HasImax                    *bool          `json:"hasImax"`
	Has3D                      *bool          `json:"has3D"`
	LastSync                   string         `json:"lastSync"`
	Countries                  []Country      `json:"countries"`
	Genres                     []Genre        `json:"genres"`
	StartYear                  *int           `json:"startYear"`
	EndYear                    *int           `json:"endYear"`
	Serial                     *bool          `json:"serial"`
	ShortFilm                  *bool          `json:"shortFilm"`
	Completed                  *bool          `json:"completed"`
	BoxOffice                  []BoxOffice    `json:"boxOffice"`
	Distributions              []Distribution `json:"distributions"`
	Facts                      []Fact         `json:"facts"`
	Seasons                    []Season       `json:"seasons"`
}

// KinopoiskFilmSearchResponse - ответ на поиск фильмов по фильтрам
// GET /api/v2.2/films
type KinopoiskFilmSearchResponse struct {
	Total      int                       `json:"total"`
	TotalPages int                       `json:"totalPages"`
	Items      []KinopoiskFilmSearchItem `json:"items"`
}

// KinopoiskFilmSearchItem - элемент поиска фильма (упрощенная версия)
type KinopoiskFilmSearchItem struct {
	FilmID            int64     `json:"filmId"`
	NameRU            *string   `json:"nameRu"`
	NameEN            *string   `json:"nameEn"`
	NameOriginal      *string   `json:"nameOriginal"`
	PosterURL         string    `json:"posterUrl"`
	PosterURLPreview  string    `json:"posterUrlPreview"`
	CoverURL          *string   `json:"coverUrl"`
	Year              *int      `json:"year"`
	FilmLength        *int      `json:"filmLength"`
	RatingKinopoisk   *float64  `json:"ratingKinopoisk"`
	RatingIMDB        *float64  `json:"ratingImdb"`
	RatingFilmCritics *float64  `json:"ratingFilmCritics"`
	RatingAwait       *float64  `json:"ratingAwait"`
	KinopoiskID       int64     `json:"kinopoiskId"`
	KinopoiskHDID     *string   `json:"kinopoiskHDId"`
	IMDBID            *string   `json:"imdbId"`
	Type              string    `json:"type"`
	Countries         []Country `json:"countries"`
	Genres            []Genre   `json:"genres"`
}

// KinopoiskSimilarFilm - похожий фильм
type KinopoiskSimilarFilm struct {
	FilmID           int64    `json:"filmId"`
	NameRU           *string  `json:"nameRu"`
	NameEN           *string  `json:"nameEn"`
	NameOriginal     *string  `json:"nameOriginal"`
	PosterURL        string   `json:"posterUrl"`
	PosterURLPreview string   `json:"posterUrlPreview"`
	RatingKinopoisk  *float64 `json:"ratingKinopoisk"`
	RatingIMDB       *float64 `json:"ratingImdb"`
	Year             *int     `json:"year"`
	Type             string   `json:"type"`
}

// KinopoiskSimilarFilmResponse - ответ с похожими фильмами
// GET /api/v2.2/films/{id}/similars
type KinopoiskSimilarFilmResponse struct {
	Total int                    `json:"total"`
	Items []KinopoiskSimilarFilm `json:"items"`
}

// KinopoiskFilmCollectionItem - элемент коллекции фильмов
// GET /api/v2.2/films/collections
type KinopoiskFilmCollectionItem struct {
	KinopoiskID     int64     `json:"kinopoiskId"`
	NameRU          *string   `json:"nameRu"`
	NameEN          *string   `json:"nameEn"`
	NameOriginal    *string   `json:"nameOriginal"`
	Countries       []Country `json:"countries"`
	Genres          []Genre   `json:"genres"`
	RatingKinopoisk *float64  `json:"ratingKinopoisk"`
	RatingImbd      *float64  `json:"ratingImbd"`
	Year            *int      `json:"year"`
	Type            string    `json:"type"`
	PosterURL       string    `json:"posterUrl"`
	PremiereRu      *string   `json:"premiereRu"`
}

// KinopoiskFilmCollectionResponse - ответ с коллекцией фильмов
// GET /api/v2.2/films/collections
type KinopoiskFilmCollectionResponse struct {
	Total      int                           `json:"total"`
	TotalPages int                           `json:"totalPages"`
	Items      []KinopoiskFilmCollectionItem `json:"items"`
}

// PremiereResponseItem - элемент премьеры
// GET /api/v2.2/films/premieres
type PremiereResponseItem struct {
	KinopoiskID      int64     `json:"kinopoiskId"`
	NameRU           *string   `json:"nameRu"`
	NameEN           *string   `json:"nameEn"`
	Year             *int      `json:"year"`
	PosterURL        string    `json:"posterUrl"`
	PosterURLPreview string    `json:"posterUrlPreview"`
	Countries        []Country `json:"countries"`
	Genres           []Genre   `json:"genres"`
	Duration         *int      `json:"duration"`
	PremiereRU       string    `json:"premiereRu"`
}

// PremiereResponse - ответ с премьерами
// GET /api/v2.2/films/premieres
type PremiereResponse struct {
	Total int                    `json:"total"`
	Items []PremiereResponseItem `json:"items"`
}
