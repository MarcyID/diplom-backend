package handlers

import (
	"diplomM/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MovieHandler struct {
	kinopoisk *service.KinopoiskClient
}

func NewMovieHandler(client *service.KinopoiskClient) *MovieHandler {
	return &MovieHandler{kinopoisk: client}
}

// SearchMovies - поиск фильмов
// GET /api/v1/movies/search?query=...&genre=...&year_from=...&year_to=...&rating_min=...&rating_max=...&page=...
func (h *MovieHandler) SearchMovies(c *gin.Context) {
	query := c.Query("query")
	genres := c.QueryArray("genre")
	countries := c.QueryArray("country")

	yearFrom, _ := strconv.Atoi(c.Query("year_from"))
	yearTo, _ := strconv.Atoi(c.Query("year_to"))

	ratingMin, _ := strconv.ParseFloat(c.Query("rating_min"), 64)
	ratingMax, _ := strconv.ParseFloat(c.Query("rating_max"), 64)

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	params := service.SearchParams{
		Keyword:   query,
		Genres:    genres,
		Countries: countries,
		YearFrom:  yearFrom,
		YearTo:    yearTo,
		RatingMin: ratingMin,
		RatingMax: ratingMax,
		Page:      page,
		Order:     "RATING",
		SortType:  "DESC",
	}

	result, err := h.kinopoisk.SearchMovies(params)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetRandomMovie - получение случайного фильма
// GET /api/v1/movies/random?genre=4&genre=17&min_rating=7.0
// genre: ID жанра (можно несколько), min_rating: минимальный рейтинг
func (h *MovieHandler) GetRandomMovie(c *gin.Context) {
	genres := c.QueryArray("genre")
	minRating, _ := strconv.ParseFloat(c.Query("min_rating"), 64)

	movie, err := h.kinopoisk.GetRandomMovie(genres, minRating)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movie)
}

// GetMovieByID - получить фильм по ID
// GET /api/v1/movies/:id
func (h *MovieHandler) GetMovieByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movie ID is required"})
		return
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	movie, err := h.kinopoisk.GetFilmByID(idInt)
	if err != nil {
		if err.Error() == "film not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movie)
}

// GetSimilarMovies - похожие фильмы
// GET /api/v1/movies/:id/similar
func (h *MovieHandler) GetSimilarMovies(c *gin.Context) {
	movieIDStr := c.Param("id")
	if movieIDStr == "" {
		movieIDStr = c.Query("movie_id")
	}

	movieID, err := strconv.ParseInt(movieIDStr, 10, 64)
	if err != nil || movieID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie_id"})
		return
	}

	result, err := h.kinopoisk.GetSimilarMovies(movieID, 0)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// FindSimilarByTitle - найти похожие по названию
// GET /api/v1/movies/similar/by-title?title=...
func (h *MovieHandler) FindSimilarByTitle(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	// Шаг 1: ищем фильм по названию
	searchParams := service.SearchParams{
		Keyword: title,
		Limit:   1,
		Page:    1,
	}

	searchResult, err := h.kinopoisk.SearchMovies(searchParams)
	if err != nil || len(searchResult.Items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	movieID := searchResult.Items[0].FilmID

	// Шаг 2: получаем похожие фильмы
	similar, err := h.kinopoisk.GetSimilarMovies(movieID, 0)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source_movie": searchResult.Items[0],
		"similar":      similar,
	})
}

// GetStaffByFilmID - получить актёров и режиссёров фильма
// GET /api/v1/movies/:id/staff
func (h *MovieHandler) GetStaffByFilmID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movie ID is required"})
		return
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	staff, err := h.kinopoisk.GetStaffByFilmID(idInt)
	if err != nil {
		if err.Error() == "staff not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
			return
		}
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, staff)
}

// GetPersonByID - получить данные о персоне (актёр, режиссёр)
// GET /api/v1/persons/:id
func (h *MovieHandler) GetPersonByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person ID is required"})
		return
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person ID"})
		return
	}

	person, err := h.kinopoisk.GetPersonByID(idInt)
	if err != nil {
		if err.Error() == "person not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, person)
}

// GetPopularFilms - популярные фильмы
// GET /api/v1/films/popular?page=1
func (h *MovieHandler) GetPopularFilms(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	result, err := h.kinopoisk.GetFilmCollections("TOP_POPULAR_MOVIES", page)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUpcomingFilms - предстоящие премьеры
// GET /api/v1/films/upcoming?page=1
func (h *MovieHandler) GetUpcomingFilms(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	result, err := h.kinopoisk.GetFilmCollections("CLOSES_RELEASES", page)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SearchActors - поиск актёров по имени
// GET /api/v1/actors/search?q=...&page=1
func (h *MovieHandler) SearchActors(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	result, err := h.kinopoisk.SearchPersons(query, page)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPremieres - получение премьер за указанный год и месяц
// GET /api/v1/films/premieres?year=2025&month=JANUARY
func (h *MovieHandler) GetPremieres(c *gin.Context) {
	yearStr := c.Query("year")
	month := c.Query("month")

	if yearStr == "" || month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year and month parameters are required"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year format"})
		return
	}

	result, err := h.kinopoisk.GetPremieres(year, month)
	if err != nil {
		if apiErr, ok := err.(*service.APIError); ok {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
