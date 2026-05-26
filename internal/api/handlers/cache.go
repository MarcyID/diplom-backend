package handlers

import (
	"diplomM/internal/cache"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CacheHandler обработчики для отладки кеша
type CacheHandler struct {
	cacheInstance cache.Cache
}

// NewCacheHandler создает новый CacheHandler
func NewCacheHandler(cacheInstance cache.Cache) *CacheHandler {
	return &CacheHandler{
		cacheInstance: cacheInstance,
	}
}

// CacheStats - статистика кеша (для отладки)
// GET /api/v1/cache/stats
func (h *CacheHandler) CacheStats(c *gin.Context) {
	// Для memory cache можно показать размер
	// Для Redis - можно сделать INFO
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Cache is working",
		"note":    "In-memory cache stats not available - check logs for HIT/MISS",
	})
}

// CacheClear - очистка кеша (для тестирования)
// POST /api/v1/cache/clear
func (h *CacheHandler) CacheClear(c *gin.Context) {
	if h.cacheInstance == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cache not initialized"})
		return
	}

	err := h.cacheInstance.Clear(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache cleared"})
}
