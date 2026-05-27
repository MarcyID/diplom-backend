package handlers

import (
	"diplomM/internal/api/response"
	"diplomM/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfileHandler обработчики для профиля пользователя
type ProfileHandler struct {
	authService       *service.AuthService
	userService       *service.UserService
	fileUploadService *service.FileUploadService
}

// NewProfileHandler создает новый ProfileHandler
func NewProfileHandler(authService *service.AuthService, userService *service.UserService, fileUploadService *service.FileUploadService) *ProfileHandler {
	return &ProfileHandler{
		authService:       authService,
		userService:       userService,
		fileUploadService: fileUploadService,
	}
}

// UpdateProfile обновляет профиль пользователя
// PUT /api/v1/profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
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

	var req struct {
		FullName  *string `json:"full_name,omitempty"`
		AvatarURL *string `json:"avatar_url,omitempty"`
		BannerURL *string `json:"banner_url,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	// Получаем текущего пользователя
	user, err := h.authService.GetUserByID(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "User not found")
		return
	}

	// Обновляем поля
	if req.FullName != nil {
		user.FullName = req.FullName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.BannerURL != nil {
		user.BannerURL = req.BannerURL
	}

	// Сохраняем изменения
	err = h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"user": user.ToUserInfo(),
	})
}

// GetProfile получает профиль пользователя
// GET /api/v1/profile/me
func (h *ProfileHandler) GetProfile(c *gin.Context) {
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

	user, err := h.authService.GetUserByID(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "User not found")
		return
	}

	response.Success(c, gin.H{
		"user": user.ToUserInfo(),
	})
}

// GetGenrePreferences получает жанровые предпочтения пользователя
// GET /api/v1/profile/genres
func (h *ProfileHandler) GetGenrePreferences(c *gin.Context) {
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

	genreIDs, err := h.userService.GetGenrePreferences(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"genre_preferences": genreIDs,
	})
}

// UpdateGenrePreferences обновляет жанровые предпочтения пользователя
// PUT /api/v1/profile/genres
func (h *ProfileHandler) UpdateGenrePreferences(c *gin.Context) {
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

	var req struct {
		GenrePreferences []int64 `json:"genre_preferences"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	// Разрешаем nil или пустой срез
	if req.GenrePreferences == nil {
		req.GenrePreferences = []int64{}
	}

	err := h.userService.UpdateGenrePreferences(c.Request.Context(), userIDInt, req.GenrePreferences)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"genre_preferences": req.GenrePreferences,
	})
}

// UploadAvatar загружает аватар пользователя
// POST /api/v1/profile/avatar
func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
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

	// Получаем файл из формы
	file, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "No avatar file provided")
		return
	}

	// Загружаем файл
	result, err := h.fileUploadService.UploadAvatar(c.Request.Context(), file, userIDInt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	// Обновляем аватар пользователя в БД
	user, err := h.authService.GetUserByID(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "User not found")
		return
	}

	user.AvatarURL = &result.URL
	err = h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"avatar_url": result.URL,
		"user":       user.ToUserInfo(),
	})
}

// UploadBanner загружает фон профиля пользователя
// POST /api/v1/profile/banner
func (h *ProfileHandler) UploadBanner(c *gin.Context) {
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

	// Получаем файл из формы
	file, err := c.FormFile("banner")
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, "No banner file provided")
		return
	}

	// Загружаем файл
	result, err := h.fileUploadService.UploadBanner(c.Request.Context(), file, userIDInt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	// Обновляем фон пользователя в БД
	user, err := h.authService.GetUserByID(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "User not found")
		return
	}

	user.BannerURL = &result.URL
	err = h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"banner_url": result.URL,
		"user":       user.ToUserInfo(),
	})
}

// DeleteAvatar удаляет аватар пользователя
// DELETE /api/v1/profile/avatar
func (h *ProfileHandler) DeleteAvatar(c *gin.Context) {
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

	// Получаем текущего пользователя
	user, err := h.authService.GetUserByID(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "User not found")
		return
	}

	// Удаляем файл если аватар был установлен
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		err := h.fileUploadService.DeleteFile(c.Request.Context(), *user.AvatarURL)
		if err != nil {
			// Логируем ошибку но не прерываем выполнение
			// Файл может быть уже удален или перемещен
		}
	}

	// Очищаем аватар в БД
	user.AvatarURL = nil
	err = h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "Avatar deleted successfully",
		"user":    user.ToUserInfo(),
	})
}

// DeleteBanner удаляет фон профиля пользователя
// DELETE /api/v1/profile/banner
func (h *ProfileHandler) DeleteBanner(c *gin.Context) {
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

	// Получаем текущего пользователя
	user, err := h.authService.GetUserByID(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "User not found")
		return
	}

	// Удаляем файл если баннер был установлен
	if user.BannerURL != nil && *user.BannerURL != "" {
		err := h.fileUploadService.DeleteFile(c.Request.Context(), *user.BannerURL)
		if err != nil {
			// Логируем ошибку но не прерываем выполнение
			// Файл может быть уже удален или перемещен
		}
	}

	// Очищаем баннер в БД
	user.BannerURL = nil
	err = h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "Banner deleted successfully",
		"user":    user.ToUserInfo(),
	})
}
