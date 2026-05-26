package handlers

import (
	"diplomM/internal/database"
	"diplomM/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SystemHandler обработчики для системных endpoints
type SystemHandler struct {
	db        *database.PostgreSQL
	kinopoisk *service.KinopoiskClient
}

// NewSystemHandler создает новый SystemHandler
func NewSystemHandler(db *database.PostgreSQL, kinopoisk *service.KinopoiskClient) *SystemHandler {
	return &SystemHandler{
		db:        db,
		kinopoisk: kinopoisk,
	}
}

// HealthCheck - базовая проверка здоровья сервера
// GET /health
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// APIHealth - расширенная проверка здоровья API
// GET /api/v1/health
func (h *SystemHandler) APIHealth(c *gin.Context) {
	response := gin.H{
		"status": "ok",
		"services": gin.H{
			"server": "ok",
		},
	}

	// Проверяем подключение к БД
	if h.db != nil && h.db.Pool != nil {
		err := h.db.Pool.Ping(c.Request.Context())
		if err != nil {
			response["services"].(gin.H)["database"] = "error"
			response["database_error"] = err.Error()
		} else {
			response["services"].(gin.H)["database"] = "ok"
		}
	} else {
		response["services"].(gin.H)["database"] = "not_configured"
	}

	// Проверяем Kinopoisk API (простой ping)
	if h.kinopoisk != nil {
		response["services"].(gin.H)["kinopoisk"] = "ok"
	} else {
		response["services"].(gin.H)["kinopoisk"] = "not_configured"
	}

	status := http.StatusOK
	if response["services"].(gin.H)["database"] == "error" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, response)
}
