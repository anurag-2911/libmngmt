package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"libmngmt/internal/models"

	"github.com/go-redis/redis/v8"
)

// Cache interface for abstraction
type Cache interface {
	// Book operations
	GetBook(ctx context.Context, id string) (*models.Book, error)
	SetBook(ctx context.Context, book *models.Book, ttl time.Duration) error
	DeleteBook(ctx context.Context, id string) error

	// Book list operations
	GetBookList(ctx context.Context, key string) (*models.BooksListResponse, error)
	SetBookList(ctx context.Context, key string, response *models.BooksListResponse, ttl time.Duration) error
	DeleteBookListCache(ctx context.Context, pattern string) error

	// Health check
	Ping(ctx context.Context) error
	Close() error
}

// RedisCache implements Cache interface
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     20,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return &RedisCache{client: rdb}
}

// GetBook retrieves a book from cache
func (r *RedisCache) GetBook(ctx context.Context, id string) (*models.Book, error) {
	key := fmt.Sprintf("book:%s", id)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var book models.Book
	err = json.Unmarshal([]byte(data), &book)
	return &book, err
}

// SetBook stores a book in cache
func (r *RedisCache) SetBook(ctx context.Context, book *models.Book, ttl time.Duration) error {
	key := fmt.Sprintf("book:%s", book.ID.String())
	data, err := json.Marshal(book)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// DeleteBook removes a book from cache
func (r *RedisCache) DeleteBook(ctx context.Context, id string) error {
	key := fmt.Sprintf("book:%s", id)
	return r.client.Del(ctx, key).Err()
}

// GetBookList retrieves a book list from cache
func (r *RedisCache) GetBookList(ctx context.Context, key string) (*models.BooksListResponse, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var response models.BooksListResponse
	err = json.Unmarshal([]byte(data), &response)
	return &response, err
}

// SetBookList stores a book list in cache
func (r *RedisCache) SetBookList(ctx context.Context, key string, response *models.BooksListResponse, ttl time.Duration) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// DeleteBookListCache removes book list caches matching pattern
func (r *RedisCache) DeleteBookListCache(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Ping checks if Redis is available
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// NoOpCache is a cache implementation that does nothing (for when Redis is disabled)
type NoOpCache struct{}

func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (n *NoOpCache) GetBook(ctx context.Context, id string) (*models.Book, error) {
	return nil, redis.Nil
}

func (n *NoOpCache) SetBook(ctx context.Context, book *models.Book, ttl time.Duration) error {
	return nil
}

func (n *NoOpCache) DeleteBook(ctx context.Context, id string) error {
	return nil
}

func (n *NoOpCache) GetBookList(ctx context.Context, key string) (*models.BooksListResponse, error) {
	return nil, redis.Nil
}

func (n *NoOpCache) SetBookList(ctx context.Context, key string, response *models.BooksListResponse, ttl time.Duration) error {
	return nil
}

func (n *NoOpCache) DeleteBookListCache(ctx context.Context, pattern string) error {
	return nil
}

func (n *NoOpCache) Ping(ctx context.Context) error {
	return nil
}

func (n *NoOpCache) Close() error {
	return nil
}
