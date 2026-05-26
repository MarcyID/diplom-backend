package handlers

import (
	"diplomM/internal/api/response"
	"diplomM/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfileHandler обработчики для профиля пользователя
type ProfileHandler struct {
	authService *service.AuthService
}

// NewProfileHandler создает новый ProfileHandler
func NewProfileHandler(authService *service.AuthService) *ProfileHandler {
	return &ProfileHandler{
		authService: authService,
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
