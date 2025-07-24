package cache

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"libmngmt/internal/models"
)

// GenerateBookListKey creates a cache key for book list queries
func GenerateBookListKey(filter models.BookFilter) string {
	var availableStr string
	if filter.Available != nil {
		availableStr = strconv.FormatBool(*filter.Available)
	}

	// Create a string representation of the filter
	filterStr := fmt.Sprintf("author:%s|genre:%s|language:%s|available:%s|limit:%d|offset:%d",
		filter.Author,
		filter.Genre,
		filter.Language,
		availableStr,
		filter.Limit,
		filter.Offset,
	)

	// Generate MD5 hash to create consistent, shorter keys
	hash := md5.Sum([]byte(filterStr))
	return fmt.Sprintf("books:%x", hash)
}

// Cache key constants
const (
	BookKeyPrefix     = "book:"
	BookListKeyPrefix = "books:"
	BookListPattern   = "books:*"
)

// Cache TTL constants
const (
	BookTTL     = 10 * 60 // 10 minutes
	BookListTTL = 5 * 60  // 5 minutes
)
