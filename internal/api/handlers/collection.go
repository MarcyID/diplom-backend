package handlers

import (
	"diplomM/internal/api/response"
	"diplomM/internal/model/collection"
	"diplomM/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CollectionHandler обработчики для подборок
type CollectionHandler struct {
	collectionService *service.CollectionService
}

// NewCollectionHandler создает новый CollectionHandler
func NewCollectionHandler(collectionService *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
	}
}

// CreateCollection создает новую подборку
// POST /api/v1/collections
func (h *CollectionHandler) CreateCollection(c *gin.Context) {
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

	var req collection.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	coll, err := h.collectionService.CreateCollection(c.Request.Context(), userIDInt, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"collection": coll,
	})
}

// GetCollection получает подборку по ID
// GET /api/v1/collections/:id
func (h *CollectionHandler) GetCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid collection ID")
		return
	}

	// Получаем userID из контекста (может не быть, если запрос публичный)
	var requestUserID *int64
	if userID, exists := c.Get("userID"); exists {
		if userIDInt, ok := userID.(int64); ok {
			requestUserID = &userIDInt
		}
	}

	collWithFilms, err := h.collectionService.GetCollection(c.Request.Context(), id, requestUserID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, err.Error())
			return
		}
		if err.Error() == "collection not found" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"collection": collWithFilms,
	})
}

// GetUserCollections получает все подборки текущего пользователя
// GET /api/v1/collections/my
func (h *CollectionHandler) GetUserCollections(c *gin.Context) {
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

	collections, total, err := h.collectionService.GetUserCollections(c.Request.Context(), userIDInt, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total": total,
		"page":  page,
		"items": collections,
	})
}

// GetPublicUserCollections получает публичные подборки пользователя
// GET /api/v1/users/:id/collections
func (h *CollectionHandler) GetPublicUserCollections(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid user ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	collections, total, err := h.collectionService.GetPublicUserCollections(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total": total,
		"page":  page,
		"items": collections,
	})
}

// UpdateCollection обновляет подборку
// PUT /api/v1/collections/:id
func (h *CollectionHandler) UpdateCollection(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid collection ID")
		return
	}

	var req collection.UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	coll, err := h.collectionService.UpdateCollection(c.Request.Context(), id, userIDInt, req)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, err.Error())
			return
		}
		if err.Error() == "collection not found" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"collection": coll,
	})
}

// DeleteCollection удаляет подборку
// DELETE /api/v1/collections/:id
func (h *CollectionHandler) DeleteCollection(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid collection ID")
		return
	}

	err = h.collectionService.DeleteCollection(c.Request.Context(), id, userIDInt)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, err.Error())
			return
		}
		if err.Error() == "collection not found" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Collection deleted successfully", nil)
}

// AddFilmToCollection добавляет фильм в подборку
// POST /api/v1/collections/:id/films
func (h *CollectionHandler) AddFilmToCollection(c *gin.Context) {
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

	idStr := c.Param("id")
	collectionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid collection ID")
		return
	}

	var req collection.AddFilmToCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	err = h.collectionService.AddFilmToCollection(c.Request.Context(), collectionID, userIDInt, req)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, err.Error())
			return
		}
		if err.Error() == "collection not found" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		if err.Error() == "film already in collection" {
			response.Error(c, http.StatusConflict, response.ErrConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Film added to collection successfully", nil)
}

// RemoveFilmFromCollection удаляет фильм из подборки
// DELETE /api/v1/collections/:id/films/:filmId
func (h *CollectionHandler) RemoveFilmFromCollection(c *gin.Context) {
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

	idStr := c.Param("id")
	collectionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid collection ID")
		return
	}

	filmIdStr := c.Param("filmId")
	filmID, err := strconv.ParseInt(filmIdStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid film ID")
		return
	}

	err = h.collectionService.RemoveFilmFromCollection(c.Request.Context(), collectionID, userIDInt, filmID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, err.Error())
			return
		}
		if err.Error() == "collection not found" || err.Error() == "film not found in collection" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Film removed from collection successfully", nil)
}

// ReorderCollectionFilms изменяет порядок фильмов в подборке
// PUT /api/v1/collections/:id/films/reorder
func (h *CollectionHandler) ReorderCollectionFilms(c *gin.Context) {
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

	idStr := c.Param("id")
	collectionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "Invalid collection ID")
		return
	}

	var req struct {
		FilmPositions map[int64]int `json:"film_positions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	err = h.collectionService.ReorderCollectionFilms(c.Request.Context(), collectionID, userIDInt, req.FilmPositions)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, err.Error())
			return
		}
		if err.Error() == "collection not found" {
			response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Films reordered successfully", nil)
}
