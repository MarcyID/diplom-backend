package service

import (
	"diplomM/internal/model"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type PoiskKinoClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewPoiskKinoClient(apiKey, baseURL string) *PoiskKinoClient {
	return &PoiskKinoClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{},
	}
}

type SearchParams struct {
	Query     string
	Genres    []string
	YearFrom  int
	YearTo    int
	RatingMin float64
	RatingMax float64
	Limit     int
}

func (c *PoiskKinoClient) SearchMovies(params SearchParams) (*model.MovieSearchResponse, error) {
	reqURL := fmt.Sprintf("%s/v1.4/movie/search", c.baseURL)
	q := url.Values{}

	if params.Query != "" {
		q.Set("query", params.Query)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
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
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var result model.MovieSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *PoiskKinoClient) GetRandomMovie(genres []string, minRating float64) (*model.Movie, error) {
	reqURL := fmt.Sprintf("%s/v1.4/movie/random", c.baseURL)
	q := url.Values{}

	for _, g := range genres {
		q.Add("genres.name", g)
	}
	if minRating > 0 {
		q.Set("rating.kp", fmt.Sprintf("%.1f-10", minRating))
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

	var movie model.Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, err
	}

	return &movie, nil
}

// Добавляем метод для получения похожих фильмов
func (c *PoiskKinoClient) GetSimilarMovies(movieID int64, limit int) ([]model.MoviePreview, error) {
	// Сначала получаем детали фильма с полем similarMovies
	reqURL := fmt.Sprintf("%s/v1.4/movie/%d", c.baseURL, movieID)

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
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var movie model.Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, err
	}

	// Ограничиваем количество похожих фильмов
	similar := movie.SimilarMovies
	if limit > 0 && len(similar) > limit {
		similar = similar[:limit]
	}

	return similar, nil
}

// Метод для получения расширенной информации о похожих фильмах
// (дополнительные запросы для получения постеров, рейтингов и т.д.)
func (c *PoiskKinoClient) GetSimilarMoviesDetailed(movieID int64, limit int) ([]model.Movie, error) {
	previews, err := c.GetSimilarMovies(movieID, limit)
	if err != nil {
		return nil, err
	}

	var detailed []model.Movie
	for _, preview := range previews {
		// Делаем запрос для каждого похожего фильма
		movie, err := c.GetMovieByID(preview.ID)
		if err == nil { // игнорируем ошибки для отдельных фильмов
			detailed = append(detailed, *movie)
		}
	}

	return detailed, nil
}

// Вспомогательный метод для получения фильма по ID
func (c *PoiskKinoClient) GetMovieByID(id int64) (*model.Movie, error) {
	reqURL := fmt.Sprintf("%s/v1.4/movie/%d", c.baseURL, id)

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
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var movie model.Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, err
	}

	return &movie, nil
}
