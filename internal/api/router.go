package api

import (
	"diplomM/internal/api/handlers"
	"diplomM/internal/api/middleware"
	"diplomM/internal/cache"
	"diplomM/internal/database"
	"diplomM/internal/repository"
	"diplomM/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

// allowedOrigins проверяет, разрешен ли origin
func allowedOrigins() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Получаем список разрешенных origins из env
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
		if allowedOriginsEnv == "" {
			// Если не настроено, разрешаем все (для backward compatibility)
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			// Проверяем, есть ли origin в списке разрешенных
			allowedOrigins := strings.Split(allowedOriginsEnv, ",")
			for _, allowed := range allowedOrigins {
				if origin == strings.TrimSpace(allowed) {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}

			// Если origin не найден в списке, не устанавливаем заголовок
			// (запрос будет заблокирован браузером)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-KEY, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 часа

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetupRouter(kinopoisk *service.KinopoiskClient, authService *service.AuthService, db *database.PostgreSQL, kinopoiskCache cache.Cache) *gin.Engine {
	r := gin.Default()

	// CORS для фронтенда
	r.Use(allowedOrigins())

	// Recovery middleware для обработки паник
	r.Use(gin.Recovery())

	movieHandler := handlers.NewMovieHandler(kinopoisk)
	authHandler := handlers.NewAuthHandler(authService)
	profileHandler := handlers.NewProfileHandler(authService)
	systemHandler := handlers.NewSystemHandler(db, kinopoisk)
	cacheHandler := handlers.NewCacheHandler(kinopoiskCache)
	collectionHandler := handlers.NewCollectionHandler(service.NewCollectionService(
		repository.NewPostgresCollectionRepository(db),
		kinopoisk,
	))
	favoriteHandler := handlers.NewFavoriteHandler(service.NewFavoriteService(
		repository.NewPostgresFavoriteRepository(db),
		kinopoisk,
	))

	// Root endpoint - информация о сервере
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "Diplom Backend API",
			"version":     "1.0.0",
			"description": "Proxy server for Kinopoisk API with user authentication",
			"endpoints": gin.H{
				"health":   "/health",
				"api":      "/api/v1",
				"api_docs": "/api/v1/health",
				"cache":    "/api/v1/cache/stats",
			},
		})
	})

	// Health check - базовая проверка
	r.GET("/health", systemHandler.HealthCheck)

	api := r.Group("/api/v1")
	{
		// Health check API - расширенная проверка
		api.GET("/health", systemHandler.APIHealth)

		// Cache endpoints (для отладки)
		api.GET("/cache/stats", cacheHandler.CacheStats)
		api.POST("/cache/clear", cacheHandler.CacheClear)

		// Информация о доступных endpoints
		api.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version": "1.0.0",
				"endpoints": gin.H{
					"auth": gin.H{
						"register": "POST /api/v1/auth/register",
						"login":    "POST /api/v1/auth/login",
						"logout":   "POST /api/v1/auth/logout (requires token)",
						"me":       "GET /api/v1/auth/me (requires token)",
						"refresh":  "POST /api/v1/auth/refresh",
					},
					"profile": gin.H{
						"get":    "GET /api/v1/profile/me (requires token)",
						"update": "PUT /api/v1/profile (requires token)",
					},
					"films": gin.H{
						"search":    "GET /api/v1/movies/search",
						"by_id":     "GET /api/v1/movies/:id",
						"random":    "GET /api/v1/movies/random",
						"popular":   "GET /api/v1/films/popular",
						"upcoming":  "GET /api/v1/films/upcoming",
						"premieres": "GET /api/v1/films/premieres",
						"similar":   "GET /api/v1/movies/:id/similar",
					},
					"actors": gin.H{
						"search": "GET /api/v1/actors/search",
						"by_id":  "GET /api/v1/actors/:id",
					},
					"collections": gin.H{
						"list":        "GET /api/v1/collections/my (requires token)",
						"get":         "GET /api/v1/collections/:id (public or owner)",
						"create":      "POST /api/v1/collections (requires token)",
						"update":      "PUT /api/v1/collections/:id (requires token, owner)",
						"delete":      "DELETE /api/v1/collections/:id (requires token, owner)",
						"add_film":    "POST /api/v1/collections/:id/films (requires token, owner)",
						"remove_film": "DELETE /api/v1/collections/:id/films/:filmId (requires token, owner)",
						"reorder":     "PUT /api/v1/collections/:id/films/reorder (requires token, owner)",
						"user_public": "GET /api/v1/users/:id/collections (public user collections)",
					},
					"favorites": gin.H{
						"list":          "GET /api/v1/favorites (requires token)",
						"add_film":      "POST /api/v1/favorites/film/:filmId (requires token)",
						"remove_film":   "DELETE /api/v1/favorites/film/:filmId (requires token)",
						"add_person":    "POST /api/v1/favorites/person/:personId (requires token)",
						"remove_person": "DELETE /api/v1/favorites/person/:personId (requires token)",
						"toggle_film":   "POST /api/v1/favorites/toggle/film/:filmId (requires token)",
						"toggle_person": "POST /api/v1/favorites/toggle/person/:personId (requires token)",
					},
					"reference": gin.H{
						"genres": "GET /api/v1/genres",
					},
				},
			})
		})

		// 🔐 Аутентификация (публичные endpoints)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)

		// 🔐 Аутентификация (требуют токен)
		auth := api.Group("")
		auth.Use(middleware.JWTAuth(authService))
		{
			auth.GET("/auth/me", authHandler.Me)
			auth.POST("/auth/logout", authHandler.Logout)
		}

		// Фильмы
		api.GET("/movies/search", movieHandler.SearchMovies)
		api.GET("/movies/random", movieHandler.GetRandomMovie)
		api.GET("/movies/:id", movieHandler.GetMovieByID)

		// 🆕 Популярные и предстоящие фильмы
		api.GET("/films/popular", movieHandler.GetPopularFilms)
		api.GET("/films/upcoming", movieHandler.GetUpcomingFilms)

		// 🆕 Похожие фильмы
		api.GET("/movies/:id/similar", movieHandler.GetSimilarMovies)        // по ID
		api.GET("/movies/similar/by-title", movieHandler.FindSimilarByTitle) // по названию (удобнее для фронта)

		// 🆕 Актёры и режиссёры
		api.GET("/movies/:id/staff", movieHandler.GetStaffByFilmID) // актёры фильма
		api.GET("/persons/:id", movieHandler.GetPersonByID)         // данные персоны

		// Актёры и персоны
		api.GET("/actors/search", movieHandler.SearchActors) // поиск актёров по имени
		api.GET("/actors/:id", movieHandler.GetPersonByID)   // данные актёра по ID (алиас на persons/:id)

		// Жанры и страны из Kinopoisk API
		api.GET("/genres", func(c *gin.Context) {
			filters, err := kinopoisk.GetFilters()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, filters)
		})

		// Премьеры за текущий и следующий год
		api.GET("/films/premieres", movieHandler.GetPremieres)

		// 🎬 Подборки фильмов
		api.GET("/collections/:id", collectionHandler.GetCollection)                  // публичный доступ или владелец
		api.GET("/users/:id/collections", collectionHandler.GetPublicUserCollections) // публичные подборки пользователя

		// Подборки (требуют токен)
		collections := api.Group("/collections")
		collections.Use(middleware.JWTAuth(authService))
		{
			collections.POST("", collectionHandler.CreateCollection)                             // создать подборку
			collections.GET("/my", collectionHandler.GetUserCollections)                         // мои подборки
			collections.PUT("/:id", collectionHandler.UpdateCollection)                          // обновить подборку
			collections.DELETE("/:id", collectionHandler.DeleteCollection)                       // удалить подборку
			collections.POST("/:id/films", collectionHandler.AddFilmToCollection)                // добавить фильм
			collections.DELETE("/:id/films/:filmId", collectionHandler.RemoveFilmFromCollection) // удалить фильм
			collections.PUT("/:id/films/reorder", collectionHandler.ReorderCollectionFilms)      // изменить порядок
		}

		// Избранное (требуют токен)
		favorites := api.Group("/favorites")
		favorites.Use(middleware.JWTAuth(authService))
		{
			favorites.GET("", favoriteHandler.GetFavorites)                                 // получить все избранное
			favorites.POST("/film/:filmId", favoriteHandler.AddFilmToFavorite)              // добавить фильм
			favorites.DELETE("/film/:filmId", favoriteHandler.RemoveFilmFromFavorite)       // удалить фильм
			favorites.POST("/person/:personId", favoriteHandler.AddPersonToFavorite)        // добавить персону
			favorites.DELETE("/person/:personId", favoriteHandler.RemovePersonFromFavorite) // удалить персону
			favorites.POST("/toggle/film/:filmId", favoriteHandler.ToggleFilm)              // переключить фильм
			favorites.POST("/toggle/person/:personId", favoriteHandler.TogglePerson)        // переключить персону
		}

		// Профиль (требуют токен)
		profile := api.Group("/profile")
		profile.Use(middleware.JWTAuth(authService))
		{
			profile.GET("/me", profileHandler.GetProfile) // получить мой профиль
			profile.PUT("", profileHandler.UpdateProfile) // обновить профиль
		}
	}
	return r
}
