package main

import (
	"diplomM/internal/api"
	"diplomM/internal/service"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env not found, using system env vars")
	}

	apiKey := os.Getenv("POISKKINO_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ POISKKINO_API_KEY not set")
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

	client := service.NewPoiskKinoClient(apiKey, "https://api.poiskkino.dev")
	router := api.SetupRouter(client)

	log.Printf("🚀 Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
