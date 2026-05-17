package api

import (
	"diplomM/internal/api/handlers"
	"diplomM/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRouter(poiskKino *service.PoiskKinoClient) *gin.Engine {
	r := gin.Default()

	// CORS для фронтенда
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-KEY")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	movieHandler := handlers.NewMovieHandler(poiskKino)

	api := r.Group("/api/v1")
	{
		// Существующие роуты
		api.GET("/movies/search", movieHandler.SearchMovies)
		api.GET("/movies/random", movieHandler.GetRandomMovie)
		api.GET("/movies/:id", movieHandler.GetMovieByID)

		// 🆕 Новые роуты для похожих фильмов
		api.GET("/movies/:id/similar", movieHandler.GetSimilarMovies)        // по ID
		api.GET("/movies/similar/by-title", movieHandler.FindSimilarByTitle) // по названию (удобнее для фронта)

		api.GET("/genres", func(c *gin.Context) {
			c.JSON(200, gin.H{"genres": []string{
				"драма", "комедия", "боевик", "ужасы",
				"фантастика", "детектив", "аниме",
			}})
		})
	}
	return r
}
