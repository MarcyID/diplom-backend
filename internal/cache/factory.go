package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// NewCache создает кеш на основе конфигурации
func NewCache(cfg Config) (Cache, func(), error) {
	var cache Cache
	var cleanup func()

	switch cfg.Type {
	case "redis":
		redisCache, err := NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		if err != nil {
			log.Printf("Redis connection failed: %v, falling back to memory cache", err)
			cache = NewMemoryCache()
			cleanup = func() {}
		} else {
			cache = redisCache
			cleanup = func() {
				redisCache.Close()
			}
			log.Println("Redis cache connected")
		}

	case "memory", "":
		cache = NewMemoryCache()
		cleanup = func() {}
		log.Println("Memory cache initialized")

	default:
		return nil, nil, fmt.Errorf("unknown cache type: %s", cfg.Type)
	}

	log.Printf("Cache TTL: %v", cfg.DefaultTTL)

	return cache, cleanup, nil
}

// GenerateKey генерирует ключ кеша для запросов Kinopoisk API
func GenerateKey(endpoint string, params map[string]string) string {
	key := endpoint
	for k, v := range params {
		key += fmt.Sprintf(":%s=%s", k, v)
	}
	return "kinopoisk:" + key
}

// CacheWithResult обёртка для кеширования результатов запросов
func CacheWithResult[T any](
	ctx context.Context,
	cache Cache,
	key string,
	ttl time.Duration,
	fetch func() (*T, error),
) (*T, error) {
	// Пытаемся получить из кеша
	var result T
	data, found, err := cache.Get(ctx, key)
	if err == nil && found {
		// Десериализуем из JSON (предполагаем, что данные в JSON)
		if err := json.Unmarshal(data, &result); err == nil {
			return &result, nil
		}
	}

	// Получаем данные из источника
	resultPtr, err := fetch()
	if err != nil {
		return nil, err
	}

	// Сериализуем и сохраняем в кеш
	if data, err := json.Marshal(resultPtr); err == nil {
		_ = cache.Set(ctx, key, data, ttl)
	}

	return resultPtr, nil
}
