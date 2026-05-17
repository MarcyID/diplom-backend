package handlers

import (
	"diplomM/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MovieHandler struct {
	poiskKino *service.PoiskKinoClient
}

func NewMovieHandler(client *service.PoiskKinoClient) *MovieHandler {
	return &MovieHandler{poiskKino: client}
}

func (h *MovieHandler) SearchMovies(c *gin.Context) {
	query := c.Query("q")
	genres := c.QueryArray("genre")

	yearFrom, _ := strconv.Atoi(c.Query("year_from"))
	yearTo, _ := strconv.Atoi(c.Query("year_to"))

	ratingMin, _ := strconv.ParseFloat(c.Query("rating_min"), 64)
	ratingMax, _ := strconv.ParseFloat(c.Query("rating_max"), 64)

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	} // лимит API

	params := service.SearchParams{
		Query:     query,
		Genres:    genres,
		YearFrom:  yearFrom,
		YearTo:    yearTo,
		RatingMin: ratingMin,
		RatingMax: ratingMax,
		Limit:     limit,
	}

	// Для сложных фильтров используем v1.5/movie с cursor
	// Для простого поиска по названию — v1.4/movie/search
	var result interface{}
	var err error

	if query != "" || len(genres) == 0 {
		result, err = h.poiskKino.SearchMovies(params)
	} else {
		// Здесь можно добавить вызов /v1.5/movie с фильтрами
		// Для краткости — упрощённая логика
		result, err = h.poiskKino.SearchMovies(params)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MovieHandler) GetRandomMovie(c *gin.Context) {
	genres := c.QueryArray("genre")
	minRating, _ := strconv.ParseFloat(c.Query("min_rating"), 64)

	movie, err := h.poiskKino.GetRandomMovie(genres, minRating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) GetMovieByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movie ID is required"})
		return
	}
	idInt, err := strconv.ParseInt(id, 10, 64) // string → int64
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	movie, err := h.poiskKino.GetMovieByID(idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) GetSimilarMovies(c *gin.Context) {
	// Получаем ID фильма из query или path
	movieIDStr := c.Param("id")
	if movieIDStr == "" {
		movieIDStr = c.Query("movie_id")
	}

	movieID, err := strconv.ParseInt(movieIDStr, 10, 64)
	if err != nil || movieID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie_id"})
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	} // ограничиваем чтобы не тратить лимиты API

	detailed := c.Query("detailed") == "true"

	if detailed {
		// Расширенный режим: получаем полные данные о похожих фильмах
		movies, err := h.poiskKino.GetSimilarMoviesDetailed(movieID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"movies": movies,
			"count":  len(movies),
		})
	} else {
		// Быстрый режим: только превью (экономит запросы к API)
		previews, err := h.poiskKino.GetSimilarMovies(movieID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"movies": previews,
			"count":  len(previews),
		})
	}
}

// Дополнительный хендлер: "найти похожие по названию" (удобно для фронтенда)
func (h *MovieHandler) FindSimilarByTitle(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	// Шаг 1: ищем фильм по названию
	searchParams := service.SearchParams{
		Query: title,
		Limit: 1, // нам нужен только первый результат
	}

	searchResult, err := h.poiskKino.SearchMovies(searchParams)
	if err != nil || len(searchResult.Docs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	movieID := searchResult.Docs[0].ID

	// Шаг 2: получаем похожие фильмы
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}

	previews, err := h.poiskKino.GetSimilarMovies(movieID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source_movie": gin.H{
			"id":     searchResult.Docs[0].ID,
			"name":   searchResult.Docs[0].Name,
			"year":   searchResult.Docs[0].Year,
			"poster": searchResult.Docs[0].Poster,
		},
		"similar": previews,
		"count":   len(previews),
	})
}
