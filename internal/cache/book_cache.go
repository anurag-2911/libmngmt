package cache

import (
	"context"
	"libmngmt/internal/models"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CacheEntry represents a cached item (for in-memory fallback)
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// BookCache provides thread-safe caching with Redis backend and in-memory fallback
type BookCache struct {
	mu       sync.RWMutex
	redis    Cache                  // Redis cache interface
	inMemory map[string]*CacheEntry // Fallback in-memory cache
	ttl      time.Duration
	cleanup  chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
	stats    CacheStats
	useRedis bool
}

// CacheStats provides cache statistics
type CacheStats struct {
	Hits       int64
	Misses     int64
	Items      int
	Evictions  int64
	RedisHits  int64
	MemoryHits int64
}

// NewBookCache creates a new cache with Redis backend
func NewBookCache(ttl time.Duration, cleanupInterval time.Duration, redisCache Cache) *BookCache {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &BookCache{
		redis:    redisCache,
		inMemory: make(map[string]*CacheEntry),
		ttl:      ttl,
		cleanup:  make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
		useRedis: redisCache != nil,
	}

	// Test Redis connection
	if cache.useRedis {
		if err := cache.redis.Ping(ctx); err != nil {
			log.Printf("Redis not available, falling back to in-memory cache: %v", err)
			cache.useRedis = false
		} else {
			log.Println("Redis cache enabled")
		}
	}

	// Start cleanup goroutine for in-memory cache
	if !cache.useRedis {
		go cache.cleanupExpired(cleanupInterval)
	}

	return cache
}

// GetBook retrieves a book from cache (Redis first, then in-memory fallback)
func (c *BookCache) GetBook(id uuid.UUID) (*models.Book, bool) {
	key := id.String()

	// Try Redis first if available
	if c.useRedis {
		if book, err := c.redis.GetBook(c.ctx, key); err == nil {
			c.mu.Lock()
			c.stats.Hits++
			c.stats.RedisHits++
			c.mu.Unlock()
			return book, true
		}
	}

	// Fallback to in-memory cache
	c.mu.RLock()
	entry, exists := c.inMemory[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Item expired, remove it
		go func() {
			c.mu.Lock()
			delete(c.inMemory, key)
			c.mu.Unlock()
		}()
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	c.stats.Hits++
	c.stats.MemoryHits++
	c.mu.Unlock()

	if book, ok := entry.Data.(*models.Book); ok {
		return book, true
	}
	return nil, false
}

// SetBook stores a book in cache (Redis and in-memory)
func (c *BookCache) SetBook(book *models.Book) {
	key := book.ID.String()

	// Store in Redis if available
	if c.useRedis {
		if err := c.redis.SetBook(c.ctx, book, c.ttl); err != nil {
			log.Printf("Failed to cache book in Redis: %v", err)
		}
	}

	// Also store in in-memory cache as fallback
	c.mu.Lock()
	c.inMemory[key] = &CacheEntry{
		Data:      book,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// GetBookList retrieves a book list from cache
func (c *BookCache) GetBookList(filter models.BookFilter) (*models.BooksListResponse, bool) {
	key := GenerateBookListKey(filter)

	// Try Redis first if available
	if c.useRedis {
		if response, err := c.redis.GetBookList(c.ctx, key); err == nil {
			c.mu.Lock()
			c.stats.Hits++
			c.stats.RedisHits++
			c.mu.Unlock()
			return response, true
		}
	}

	// Fallback to in-memory cache
	c.mu.RLock()
	entry, exists := c.inMemory[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Item expired, remove it
		go func() {
			c.mu.Lock()
			delete(c.inMemory, key)
			c.mu.Unlock()
		}()
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	c.stats.Hits++
	c.stats.MemoryHits++
	c.mu.Unlock()

	if response, ok := entry.Data.(*models.BooksListResponse); ok {
		return response, true
	}
	return nil, false
}

// SetBookList stores a book list in cache
func (c *BookCache) SetBookList(filter models.BookFilter, response *models.BooksListResponse) {
	key := GenerateBookListKey(filter)

	// Store in Redis if available
	if c.useRedis {
		if err := c.redis.SetBookList(c.ctx, key, response, c.ttl); err != nil {
			log.Printf("Failed to cache book list in Redis: %v", err)
		}
	}

	// Also store in in-memory cache as fallback
	c.mu.Lock()
	c.inMemory[key] = &CacheEntry{
		Data:      response,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// InvalidateBook removes a book from cache and related book lists
func (c *BookCache) InvalidateBook(id uuid.UUID) {
	key := id.String()

	// Remove from Redis if available
	if c.useRedis {
		if err := c.redis.DeleteBook(c.ctx, key); err != nil {
			log.Printf("Failed to delete book from Redis: %v", err)
		}
		// Also invalidate book list caches
		if err := c.redis.DeleteBookListCache(c.ctx, BookListPattern); err != nil {
			log.Printf("Failed to invalidate book list cache in Redis: %v", err)
		}
	}

	// Remove from in-memory cache
	c.mu.Lock()
	delete(c.inMemory, key)

	// Remove all book list caches (simple approach for in-memory)
	for k := range c.inMemory {
		if len(k) == 32 && k != key { // Assuming book list keys are MD5 hashes (32 chars)
			delete(c.inMemory, k)
		}
	}
	c.mu.Unlock()
}

// Get retrieves an item from cache (legacy method for backward compatibility)
func (c *BookCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.inMemory[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Item expired, remove it
		go func() {
			c.mu.Lock()
			delete(c.inMemory, key)
			c.mu.Unlock()
		}()
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	c.stats.Hits++
	c.stats.MemoryHits++
	c.mu.Unlock()
	return entry.Data, true
}

// Set stores an item in cache (legacy method for backward compatibility)
func (c *BookCache) Set(key string, data interface{}) {
	c.mu.Lock()
	c.inMemory[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// Delete removes an item from cache
func (c *BookCache) Delete(key string) {
	c.mu.Lock()
	delete(c.inMemory, key)
	c.mu.Unlock()
}

// Clear removes all items from cache
func (c *BookCache) Clear() {
	c.mu.Lock()
	// Clear Redis if available
	if c.useRedis {
		// Note: In production, you might want to use FLUSHDB with caution
		// For now, we'll just clear book-related keys
		go func() {
			ctx := context.Background()
			c.redis.DeleteBookListCache(ctx, "book:*")
			c.redis.DeleteBookListCache(ctx, "books:*")
		}()
	}

	// Clear in-memory cache
	for key := range c.inMemory {
		delete(c.inMemory, key)
	}
	c.mu.Unlock()
}

// Size returns the number of items in the in-memory cache
func (c *BookCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.inMemory)
}

// Stats returns cache statistics
func (c *BookCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Items = len(c.inMemory)
	return stats
}

// Shutdown gracefully shuts down the cache
func (c *BookCache) Shutdown() {
	c.cancel()
	close(c.cleanup)

	if c.useRedis && c.redis != nil {
		c.redis.Close()
	}
}

// Info returns cache information
func (c *BookCache) Info() map[string]interface{} {
	stats := c.Stats()

	return map[string]interface{}{
		"type":        "hybrid", // Redis + in-memory
		"redis":       c.useRedis,
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"redis_hits":  stats.RedisHits,
		"memory_hits": stats.MemoryHits,
		"hit_rate":    float64(stats.Hits) / float64(stats.Hits+stats.Misses),
		"size":        len(c.inMemory),
		"ttl":         c.ttl.String(),
	}
}

// cleanupExpired removes expired items from in-memory cache
func (c *BookCache) cleanupExpired(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			var expiredKeys []string

			for key, entry := range c.inMemory {
				if now.After(entry.ExpiresAt) {
					expiredKeys = append(expiredKeys, key)
				}
			}

			for _, key := range expiredKeys {
				delete(c.inMemory, key)
				c.stats.Evictions++
			}
			c.mu.Unlock()

		case <-c.ctx.Done():
			return
		}
	}
}
