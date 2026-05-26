package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache in-memory кеш с TTL
type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache создает новый in-memory кеш
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}

	// Запускаем фоновую очистку устаревших записей
	go cache.cleanupLoop(10 * time.Minute)

	return cache
}

// Get получает значение из кеша
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false, nil
	}

	// Проверяем, не истекло ли время жизни
	if time.Now().After(item.expiration) {
		return nil, false, nil
	}

	return item.value, true, nil
}

// Set устанавливает значение в кеш с TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	return nil
}

// Delete удаляет значение из кеша
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Clear очищает весь кеш
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheItem)
	return nil
}

// cleanupLoop периодически удаляет устаревшие записи
func (c *MemoryCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup удаляет устаревшие записи
func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expiration) {
			delete(c.data, key)
		}
	}
}
