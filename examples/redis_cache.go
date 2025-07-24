// Example Redis integration for caching
// This shows how Redis could be integrated for performance improvements
// Run: go mod tidy && go get github.com/go-redis/redis/v8

package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"
	// "github.com/go-redis/redis/v8" // Uncomment when implementing
)

// Example cache interface
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

// MockRedisCache for demonstration
type MockRedisCache struct {
	data map[string]string
}

func NewMockRedisCache() *MockRedisCache {
	return &MockRedisCache{
		data: make(map[string]string),
	}
}

func (m *MockRedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.data[key] = string(data)
	return nil
}

func (m *MockRedisCache) Get(ctx context.Context, key string) (string, error) {
	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("key not found")
}

func (m *MockRedisCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// BookFilter represents search filters
type BookFilter struct {
	Author    string
	Genre     string
	Language  string
	Available *bool
	Limit     int
	Offset    int
}

// CacheKey generates a unique cache key for the filter
func (f BookFilter) CacheKey() string {
	data := fmt.Sprintf("%s:%s:%s:%v:%d:%d",
		f.Author, f.Genre, f.Language, f.Available, f.Limit, f.Offset)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// BookService with caching
type BookServiceWithCache struct {
	cache Cache
	// db    repository.BookRepository // Original service
}

func NewBookServiceWithCache(cache Cache) *BookServiceWithCache {
	return &BookServiceWithCache{
		cache: cache,
	}
}

// Example cached method
func (s *BookServiceWithCache) GetBookByID(ctx context.Context, id string) (interface{}, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("book:%s", id)

	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		fmt.Printf("Cache HIT for book:%s\n", id)
		return cached, nil
	}

	fmt.Printf("Cache MISS for book:%s\n", id)

	// If not in cache, get from database
	// book, err := s.db.GetByID(id)
	// if err != nil { return nil, err }

	// Simulate database response
	book := map[string]interface{}{
		"id":     id,
		"title":  "Example Book",
		"author": "Example Author",
	}

	// Cache the result
	s.cache.Set(ctx, cacheKey, book, 10*time.Minute)

	return book, nil
}

// Example usage
func main() {
	fmt.Println("Redis Cache Integration Example")
	fmt.Println("===============================")

	cache := NewMockRedisCache()
	service := NewBookServiceWithCache(cache)
	ctx := context.Background()

	// First call - cache miss
	book1, _ := service.GetBookByID(ctx, "123")
	fmt.Printf("Result 1: %v\n", book1)

	// Second call - cache hit
	book2, _ := service.GetBookByID(ctx, "123")
	fmt.Printf("Result 2: %v\n", book2)

	// Filter cache key example
	filter := BookFilter{
		Author: "tolkien",
		Genre:  "fantasy",
		Limit:  10,
		Offset: 0,
	}
	fmt.Printf("Cache key for filter: books:%s\n", filter.CacheKey())
}

// Performance improvement estimates with Redis:
//
// Without Redis:
// - Average response time: 50-200ms
// - Database queries per request: 1-3
// - Max throughput: ~1,000 req/sec
//
// With Redis:
// - Cached response time: 1-5ms (10-50x faster)
// - Cache hit rate: 70-90% (depending on traffic patterns)
// - Max throughput: ~10,000+ req/sec
// - Database load reduction: 70-90%
//
// Implementation steps:
// 1. Add Redis to docker-compose.yml
// 2. Install Redis client: go get github.com/go-redis/redis/v8
// 3. Implement cache layer in service
// 4. Add cache invalidation on updates/deletes
// 5. Monitor cache hit rates and adjust TTL values
