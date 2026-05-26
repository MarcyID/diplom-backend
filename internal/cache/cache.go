package cache

import (
	"context"
	"time"
)

// Cache интерфейс для кеширования
type Cache interface {
	// Get получает значение из кеша
	Get(ctx context.Context, key string) ([]byte, bool, error)
	// Set устанавливает значение в кеш с TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete удаляет значение из кеша
	Delete(ctx context.Context, key string) error
	// Clear очищает весь кеш (опционально)
	Clear(ctx context.Context) error
}

// Config конфигурация кеша
type Config struct {
	// Тип кеша: "redis" или "memory"
	Type string
	// TTL по умолчанию
	DefaultTTL time.Duration
	// Redis адрес (если используется Redis)
	RedisAddr string
	// Redis пароль
	RedisPassword string
	// Redis DB номер
	RedisDB int
}
