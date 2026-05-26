package handlers

import (
	"diplomM/internal/api/response"
	"diplomM/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FavoriteHandler обработчики для избранного
type FavoriteHandler struct {
	favoriteService *service.FavoriteService
}

// NewFavoriteHandler создает новый FavoriteHandler
func NewFavoriteHandler(favoriteService *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteService: favoriteService,
	}
}

// GetFavorites получает все избранные объекты пользователя
// GET /api/v1/favorites
func (h *FavoriteHandler) GetFavorites(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "Unauthorized")
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, "Invalid user ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, err := h.favoriteService.GetFavorites(c.Request.Context(), userIDInt, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total": total,
		"page":  page,
		"items": items,
	})
}

// AddFilmToFavorite добавляет фильм в избранное
// POST /api/v1/favorites/film/:filmId
func (h *FavoriteHandler) AddFilmToFavorite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "Unauthorized")
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, "Invalid user ID")
		return
	}

	filmIdStr := c.Param("filmId")
	filmID, err := strconv.ParseInt(filmIdStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid film ID")
		return
	}

	err = h.favoriteService.AddFilm(c.Request.Context(), userIDInt, filmID)
	if err != nil {
		if err.Error() == "already in favorites" {
			response.Error(c, http.StatusConflict, response.ErrConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Film added to favorites", nil)
}

// RemoveFilmFromFavorite удаляет фильм из избранного
// DELETE /api/v1/favorites/film/:filmId
func (h *FavoriteHandler) RemoveFilmFromFavorite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "Unauthorized")
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, "Invalid user ID")
		return
	}

	filmIdStr := c.Param("filmId")
	filmID, err := strconv.ParseInt(filmIdStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid film ID")
		return
	}

	err = h.favoriteService.RemoveFilm(c.Request.Context(), userIDInt, filmID)
	if err != nil {
		if err.Error() == "not in favorites" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Film removed from favorites", nil)
}

// AddPersonToFavorite добавляет персону в избранное
// POST /api/v1/favorites/person/:personId
func (h *FavoriteHandler) AddPersonToFavorite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "Unauthorized")
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, "Invalid user ID")
		return
	}

	personIdStr := c.Param("personId")
	personID, err := strconv.ParseInt(personIdStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid person ID")
		return
	}

	err = h.favoriteService.AddPerson(c.Request.Context(), userIDInt, personID)
	if err != nil {
		if err.Error() == "already in favorites" {
			response.Error(c, http.StatusConflict, response.ErrConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Person added to favorites", nil)
}

// RemovePersonFromFavorite удаляет персону из избранного
// DELETE /api/v1/favorites/person/:personId
func (h *FavoriteHandler) RemovePersonFromFavorite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "Unauthorized")
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, "Invalid user ID")
		return
	}

	personIdStr := c.Param("personId")
	personID, err := strconv.ParseInt(personIdStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid person ID")
		return
	}

	err = h.favoriteService.RemovePerson(c.Request.Context(), userIDInt, personID)
	if err != nil {
		if err.Error() == "not in favorites" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Person removed from favorites", nil)
}

// ToggleFilm переключает статус фильма в избранном
// POST /api/v1/favorites/toggle/film/:filmId
func (h *FavoriteHandler) ToggleFilm(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "Unauthorized")
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, "Invalid user ID")
		return
	}

	filmIdStr := c.Param("filmId")
	filmID, err := strconv.ParseInt(filmIdStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid film ID")
		return
	}

	added, err := h.favoriteService.ToggleFilm(c.Request.Context(), userIDInt, filmID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	if added {
		response.SuccessWithMessage(c, "Film added to favorites", gin.H{
			"in_favorites": true,
		})
	} else {
		response.SuccessWithMessage(c, "Film removed from favorites", gin.H{
			"in_favorites": false,
		})
	}
}

// TogglePerson переключает статус персоны в избранном
// POST /api/v1/favorites/toggle/person/:personId
func (h *FavoriteHandler) TogglePerson(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "Unauthorized")
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, "Invalid user ID")
		return
	}

	personIdStr := c.Param("personId")
	personID, err := strconv.ParseInt(personIdStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid person ID")
		return
	}

	added, err := h.favoriteService.TogglePerson(c.Request.Context(), userIDInt, personID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	if added {
		response.SuccessWithMessage(c, "Person added to favorites", gin.H{
			"in_favorites": true,
		})
	} else {
		response.SuccessWithMessage(c, "Person removed from favorites", gin.H{
			"in_favorites": false,
		})
	}
}
