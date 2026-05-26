package kinopoisk

// Country - страна производства
type Country struct {
	Name string `json:"country"`
}

// Genre - жанр фильма
type Genre struct {
	Name string `json:"genre"`
}

// Fact - факт или ошибка в фильме
type Fact struct {
	Text    string `json:"text"`
	Type    string `json:"type"` // FACT, BLOOPER
	Spoiler bool   `json:"spoiler"`
}

// Season - сезон сериала
type Season struct {
	SeasonNumber int    `json:"seasonNumber"`
	Year         string `json:"year"`
}

// Distribution - информация о прокате
type Distribution struct {
	Country        string `json:"country"`
	ReleaseDate    string `json:"releaseDate"`
	ReleaseCountry string `json:"releaseCountry"`
}

// BoxOffice - бюджет и сборы
type BoxOffice struct {
	Type         string `json:"type"`
	Amount       int    `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
}

// FilterItem - элемент фильтра (жанр или страна)
type FilterItem struct {
	ID      int    `json:"id"`
	Genre   string `json:"genre"`   // Для жанров
	Country string `json:"country"` // Для стран
}

// KinopoiskFiltersResponse - ответ с фильтрами (жанры и страны)
// GET /api/v2.2/films/filters
type KinopoiskFiltersResponse struct {
	Genres    []FilterItem `json:"genres"`
	Countries []FilterItem `json:"countries"`
}
