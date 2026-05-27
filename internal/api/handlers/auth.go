package handlers

import (
	"diplomM/internal/api/response"
	"diplomM/internal/model/auth"
	"diplomM/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler обработчики для аутентификации
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler создает новый AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register регистрирует нового пользователя
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusConflict, response.ErrConflict, err.Error())
		return
	}

	response.SuccessWithMessage(c, "User registered successfully", gin.H{
		"user": user.ToUserInfo(),
	})
}

// Login выполняет вход пользователя
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	// Получаем User-Agent и IP для сессии
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	authResponse, err := h.authService.Login(c.Request.Context(), req, &userAgent, &ipAddress)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, err.Error())
		return
	}

	// Возвращаем токены напрямую (без обёртки response.Success)
	c.JSON(http.StatusOK, authResponse)
}

// Logout выполняет выход пользователя
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Получаем refresh токен из тела запроса
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	err := h.authService.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Logged out successfully", nil)
}

// Me возвращает информацию о текущем пользователе
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	// Получаем userID из контекста (устанавливается middleware)
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

	userInfo, err := h.authService.GetMe(c.Request.Context(), userIDInt)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, err.Error())
		return
	}

	response.Success(c, userInfo)
}

// Refresh обновляет пару токенов
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrValidation, err.Error())
		return
	}

	authResponse, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, err.Error())
		return
	}

	response.Success(c, authResponse)
}
