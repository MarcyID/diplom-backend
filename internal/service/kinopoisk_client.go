package service

import (
	"context"
	"diplomM/internal/cache"
	"diplomM/internal/model/kinopoisk"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// KinopoiskClient - клиент для работы с Kinopoisk API Unofficial
type KinopoiskClient struct {
	baseURL  string
	apiKey   string
	client   *http.Client
	cache    cache.Cache
	cacheTTL time.Duration
}

// KinopoiskClientConfig конфигурация клиента
type KinopoiskClientConfig struct {
	APIKey   string
	BaseURL  string
	Cache    cache.Cache
	CacheTTL time.Duration
}

// createHTTPClient создает HTTP клиент с оптимизированным транспортом
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// NewKinopoiskClient создает новый экземпляр клиента Kinopoisk API
func NewKinopoiskClient(apiKey, baseURL string) *KinopoiskClient {
	return &KinopoiskClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		client:   createHTTPClient(),
		cache:    nil,
		cacheTTL: 24 * time.Hour,
	}
}

// NewKinopoiskClientWithCache создает клиента с кешированием
func NewKinopoiskClientWithCache(config KinopoiskClientConfig) *KinopoiskClient {
	client := &KinopoiskClient{
		baseURL:  config.BaseURL,
		apiKey:   config.APIKey,
		client:   createHTTPClient(),
		cache:    config.Cache,
		cacheTTL: config.CacheTTL,
	}

	if client.cacheTTL == 0 {
		client.cacheTTL = 24 * time.Hour
	}

	return client
}

// SetCache устанавливает кеш для клиента
func (c *KinopoiskClient) SetCache(cache cache.Cache, ttl time.Duration) {
	c.cache = cache
	if ttl > 0 {
		c.cacheTTL = ttl
	}
}

// SearchParams - параметры поиска фильмов
type SearchParams struct {
	Keyword   string
	Genres    []string
	Countries []string
	YearFrom  int
	YearTo    int
	RatingMin float64
	RatingMax float64
	Type      string // "FILM", "TV_SERIES", "TV_SHOW", "MINI_SERIES", "ALL"
	Order     string // "RATING", "NUM_VOTE", "YEAR"
	SortType  string // "DESC", "ASC"
	Page      int
	Limit     int
}

// SearchMovies - поиск фильмов по фильтрам через /api/v2.2/films
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) SearchMovies(params SearchParams) (*kinopoisk.KinopoiskFilmSearchResponse, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheParams := map[string]string{
		"keyword":    params.Keyword,
		"genres":     joinInts(params.Genres),
		"countries":  joinInts(params.Countries),
		"yearFrom":   strconv.Itoa(params.YearFrom),
		"yearTo":     strconv.Itoa(params.YearTo),
		"ratingFrom": strconv.FormatFloat(params.RatingMin, 'f', 1, 64),
		"ratingTo":   strconv.FormatFloat(params.RatingMax, 'f', 1, 64),
		"type":       params.Type,
		"order":      params.Order,
		"sortType":   params.SortType,
		"page":       strconv.Itoa(params.Page),
	}
	cacheKey := c.cacheKey("/films", cacheParams)

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var result kinopoisk.KinopoiskFilmSearchResponse
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return &result, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v2.2/films", c.baseURL)
	q := url.Values{}

	// Ключевое слово для поиска
	if params.Keyword != "" {
		q.Set("keyword", params.Keyword)
	}

	// Жанры (ID через запятую)
	if len(params.Genres) > 0 {
		q.Set("genres", joinInts(params.Genres))
	}

	// Страны (ID через запятую)
	if len(params.Countries) > 0 {
		q.Set("countries", joinInts(params.Countries))
	}

	// Годы
	if params.YearFrom > 0 {
		q.Set("yearFrom", strconv.Itoa(params.YearFrom))
	}
	if params.YearTo > 0 {
		q.Set("yearTo", strconv.Itoa(params.YearTo))
	}

	// Рейтинг
	if params.RatingMin > 0 {
		q.Set("ratingFrom", strconv.FormatFloat(params.RatingMin, 'f', 1, 64))
	}
	if params.RatingMax > 0 {
		q.Set("ratingTo", strconv.FormatFloat(params.RatingMax, 'f', 1, 64))
	}

	// Тип
	if params.Type != "" {
		q.Set("type", params.Type)
	}

	// Сортировка (только order, sortType вызывает 400 ошибку)
	if params.Order != "" {
		q.Set("order", params.Order)
	}
	// SortType не используется! Kinopoisk API возвращает 400 при наличии этого параметра

	// Пагинация
	if params.Page > 0 {
		q.Set("page", strconv.Itoa(params.Page))
	}

	req, err := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var result kinopoisk.KinopoiskFilmSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Сохраняем в кеш (24 часа)
	c.cacheSet(ctx, cacheKey, jsonMustMarshal(&result))

	return &result, nil
}

// GetFilmByID - получение фильма по ID через /api/v2.2/films/{id}
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) GetFilmByID(filmID int64) (*kinopoisk.KinopoiskFilm, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheKey := c.cacheKey("/films/id", map[string]string{
		"id": strconv.FormatInt(filmID, 10),
	})

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var film kinopoisk.KinopoiskFilm
		if err := json.Unmarshal(cachedData, &film); err == nil {
			return &film, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v2.2/films/%d", c.baseURL, filmID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("film not found")
	}
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var film kinopoisk.KinopoiskFilm
	if err := json.NewDecoder(resp.Body).Decode(&film); err != nil {
		return nil, err
	}

	// Сохраняем в кеш (24 часа)
	c.cacheSet(ctx, cacheKey, jsonMustMarshal(&film))

	return &film, nil
}

// GetSimilarMovies - получение похожих фильмов через /api/v2.2/films/{id}/similars
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) GetSimilarMovies(filmID int64, limit int) (*kinopoisk.KinopoiskSimilarFilmResponse, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheKey := c.cacheKey("/films/similars", map[string]string{
		"id": strconv.FormatInt(filmID, 10),
	})

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var result kinopoisk.KinopoiskSimilarFilmResponse
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return &result, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v2.2/films/%d/similars", c.baseURL, filmID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var result kinopoisk.KinopoiskSimilarFilmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Сохраняем в кеш (24 часа)
	c.cacheSet(ctx, cacheKey, jsonMustMarshal(&result))

	return &result, nil
}

// GetRandomMovie - получение случайного фильма (эмуляция через поиск)
func (c *KinopoiskClient) GetRandomMovie(genres []string, minRating float64) (*kinopoisk.KinopoiskFilm, error) {
	params := SearchParams{
		Genres:    genres,
		RatingMin: minRating,
		Type:      "ALL",
		Order:     "RATING",
		Page:      1,
		Limit:     100,
	}

	result, err := c.SearchMovies(params)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no movies found with specified filters")
	}

	// Выбираем случайный фильм из полученных результатов
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(result.Items))

	// Используем KinopoiskID вместо FilmID (API возвращает kinopoiskId, а не filmId)
	filmID := result.Items[randomIndex].KinopoiskID
	if filmID == 0 {
		filmID = result.Items[randomIndex].FilmID
	}

	// Получаем полные данные о фильме
	return c.GetFilmByID(filmID)
}

// GetFilters - получение списка жанров и стран для фильтров
func (c *KinopoiskClient) GetFilters() (*kinopoisk.KinopoiskFiltersResponse, error) {
	reqURL := fmt.Sprintf("%s/api/v2.2/films/filters", c.baseURL)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var filters kinopoisk.KinopoiskFiltersResponse
	if err := json.NewDecoder(resp.Body).Decode(&filters); err != nil {
		return nil, err
	}

	return &filters, nil
}

// GetStaffByFilmID - получение списка актёров и режиссёров фильма
// GET /api/v1/staff?filmId={id}
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) GetStaffByFilmID(filmID int64) ([]kinopoisk.KinopoiskStaffResponse, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheKey := c.cacheKey("/staff", map[string]string{
		"filmId": strconv.FormatInt(filmID, 10),
	})

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var staff []kinopoisk.KinopoiskStaffResponse
		if err := json.Unmarshal(cachedData, &staff); err == nil {
			return staff, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v1/staff", c.baseURL)
	q := url.Values{}
	q.Set("filmId", strconv.FormatInt(filmID, 10))

	req, err := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("staff not found")
	}
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var staff []kinopoisk.KinopoiskStaffResponse
	if err := json.NewDecoder(resp.Body).Decode(&staff); err != nil {
		return nil, err
	}

	// Сохраняем в кеш (24 часа)
	c.cacheSet(ctx, cacheKey, jsonMustMarshal(&staff))

	return staff, nil
}

// GetPersonByID - получение данных о персоне (актёр, режиссёр) по ID
// GET /api/v1/staff/{id}
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) GetPersonByID(personID int64) (*kinopoisk.KinopoiskPersonResponse, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheKey := c.cacheKey("/staff/id", map[string]string{
		"id": strconv.FormatInt(personID, 10),
	})

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var person kinopoisk.KinopoiskPersonResponse
		if err := json.Unmarshal(cachedData, &person); err == nil {
			return &person, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v1/staff/%d", c.baseURL, personID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("person not found")
	}
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var person kinopoisk.KinopoiskPersonResponse
	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return nil, err
	}

	return &person, nil
}

// GetFilmCollections - получение коллекции фильмов (топы, популярные и т.д.)
// GET /api/v2.2/films/collections?type=TOP_POPULAR_MOVIES&page=1
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) GetFilmCollections(collectionType string, page int) (*kinopoisk.KinopoiskFilmCollectionResponse, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheKey := c.cacheKey("/collections", map[string]string{
		"type": collectionType,
		"page": strconv.Itoa(page),
	})

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var result kinopoisk.KinopoiskFilmCollectionResponse
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return &result, nil
		}
	}

	// Запрос к API
	reqURL := fmt.Sprintf("%s/api/v2.2/films/collections", c.baseURL)
	q := url.Values{}

	if collectionType != "" {
		q.Set("type", collectionType)
	} else {
		q.Set("type", "TOP_POPULAR_ALL")
	}

	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}

	req, err := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var result kinopoisk.KinopoiskFilmCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Сохраняем в кеш
	c.cacheSet(ctx, cacheKey, jsonMustMarshal(&result))

	return &result, nil
}

// GetPremieres - получение премьер за указанный месяц и год
// GET /api/v2.2/films/premieres?year=2025&month=JANUARY
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) GetPremieres(year int, month string) (*kinopoisk.PremiereResponse, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheKey := c.cacheKey("/premieres", map[string]string{
		"year":  strconv.Itoa(year),
		"month": month,
	})

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var result kinopoisk.PremiereResponse
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return &result, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v2.2/films/premieres", c.baseURL)
	q := url.Values{}
	q.Set("year", strconv.Itoa(year))
	q.Set("month", month)

	req, err := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var result kinopoisk.PremiereResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Сохраняем в кеш
	c.cacheSet(ctx, cacheKey, jsonMustMarshal(&result))

	return &result, nil
}

// GetAllPremieres - получение премьер за указанный год и месяц
// Возвращает премьеры из API /films/premieres
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) GetAllPremieres(year int, month string) (*kinopoisk.PremiereResponse, error) {
	return c.GetPremieres(year, month)
}

// SearchPersons - поиск актёров, режиссёров по имени
// GET /api/v1/persons?name=...&page=1
// КЕШИРОВАНИЕ: 24 часа
func (c *KinopoiskClient) SearchPersons(name string, page int) (*kinopoisk.KinopoiskPersonByNameResponse, error) {
	ctx := context.Background()

	// Генерируем ключ кеша
	cacheKey := c.cacheKey("/persons", map[string]string{
		"name": name,
		"page": strconv.Itoa(page),
	})

	// Пытаемся получить из кеша
	if cachedData, found := c.cacheGet(ctx, cacheKey); found {
		var result kinopoisk.KinopoiskPersonByNameResponse
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return &result, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v1/persons", c.baseURL)
	q := url.Values{}

	if name != "" {
		q.Set("name", name)
	}

	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}

	req, err := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, NewKinopoiskError(resp.StatusCode, string(errorBody))
	}

	var result kinopoisk.KinopoiskPersonByNameResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Сохраняем в кеш (6 часов)
	c.cacheSet(ctx, cacheKey, jsonMustMarshal(&result))

	return &result, nil
}

// Вспомогательная функция
func joinInts(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += "," + strs[i]
	}
	return result
}

// cacheGet пытается получить данные из кеша
func (c *KinopoiskClient) cacheGet(ctx context.Context, key string) ([]byte, bool) {
	if c.cache == nil {
		return nil, false
	}

	data, found, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	return data, found
}

// cacheSet сохраняет данные в кеш
func (c *KinopoiskClient) cacheSet(ctx context.Context, key string, data []byte) {
	if c.cache == nil {
		return
	}

	if len(data) == 0 {
		return
	}

	_ = c.cache.Set(ctx, key, data, c.cacheTTL)
}

// cacheKey генерирует ключ кеша для endpoint
func (c *KinopoiskClient) cacheKey(endpoint string, params map[string]string) string {
	return cache.GenerateKey(endpoint, params)
}

// jsonMustMarshal сериализует данные в JSON (паникует при ошибке)
func jsonMustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte{}
	}
	return data
}
