package main

import (
	"diplomM/internal/api"
	"diplomM/internal/cache"
	"diplomM/internal/database"
	"diplomM/internal/repository"
	"diplomM/internal/service"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system env vars")
	}

	apiKey := os.Getenv("KINOPOISK_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ KINOPOISK_API_KEY not set")
	}

	// 🎯 Читаем PORT из .env с дефолтом 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Опционально: валидация, что порт — число
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("❌ Invalid PORT value: %s (must be integer)", port)
	}

	// 📦 Инициализация кеша
	cacheType := os.Getenv("CACHE_TYPE")
	if cacheType == "" {
		cacheType = "memory" // по умолчанию memory
	}

	cacheTTL := 24 * time.Hour
	if ttlStr := os.Getenv("CACHE_TTL_HOURS"); ttlStr != "" {
		if ttlHours, err := strconv.Atoi(ttlStr); err == nil && ttlHours > 0 {
			cacheTTL = time.Duration(ttlHours) * time.Hour
		}
	}

	cacheConfig := cache.Config{
		Type:          cacheType,
		DefaultTTL:    cacheTTL,
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
	}

	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		cacheConfig.RedisDB, _ = strconv.Atoi(redisDB)
	}

	kinopoiskCache, cacheCleanup, err := cache.NewCache(cacheConfig)
	if err != nil {
		log.Printf("Cache initialization failed: %v, using no cache", err)
	}
	defer cacheCleanup()

	// 🔐 Инициализация PostgreSQL
	var db *database.PostgreSQL
	var authService *service.AuthService

	dbHost := os.Getenv("DB_HOST")
	if dbHost != "" {
		dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
		if dbPort == 0 {
			dbPort = 5432
		}

		dbConfig := database.Config{
			Host:     dbHost,
			Port:     dbPort,
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		}

		var err error
		db, err = database.NewPostgreSQL(dbConfig)
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			log.Println("Auth endpoints will be disabled")
		} else {
			log.Println("Database connected")

			// Создаем репозитории
			userRepo := repository.NewUserRepository(db)
			sessionRepo := repository.NewSessionRepository(db)

			// Создаем AuthService
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				jwtSecret = "your-secret-key-change-in-production"
				log.Println("Using default JWT_SECRET - change in production!")
			}

			authService = service.NewAuthService(
				userRepo,
				sessionRepo,
				service.AuthServiceConfig{
					JWTSecretKey:    jwtSecret,
					AccessTokenExp:  15 * time.Minute,
					RefreshTokenExp: 7 * 24 * time.Hour,
				},
			)
			log.Println("Auth service initialized")
		}
	} else {
		log.Println("DB_HOST not set - auth endpoints will be disabled")
	}

	// Инициализация Kinopoisk клиента с кешированием
	client := service.NewKinopoiskClientWithCache(service.KinopoiskClientConfig{
		APIKey:   apiKey,
		BaseURL:  "https://kinopoiskapiunofficial.tech",
		Cache:    kinopoiskCache,
		CacheTTL: cacheTTL,
	})

	log.Printf("Kinopoisk cache enabled (TTL: %v)", cacheTTL)

	// Инициализация роутера
	router := api.SetupRouter(client, authService, db, kinopoiskCache)

	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
