package service

import (
	"context"
	"fmt"
	"libmngmt/internal/cache"
	"libmngmt/internal/models"
	"libmngmt/internal/repository"
	"libmngmt/internal/workers"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BookService defines the interface for book business logic
type BookService interface {
	CreateBook(req *models.CreateBookRequest) (*models.Book, error)
	GetBookByID(id uuid.UUID) (*models.Book, error)
	GetAllBooks(filter models.BookFilter) (*models.BooksListResponse, error)
	UpdateBook(id uuid.UUID, req *models.UpdateBookRequest) (*models.Book, error)
	DeleteBook(id uuid.UUID) error
	BulkCreateBooks(requests []*models.CreateBookRequest) ([]*models.Book, []error)
	GetMetrics() ServiceMetrics
	Shutdown(ctx context.Context) error
}

// ServiceMetrics tracks service performance
type ServiceMetrics struct {
	mu           sync.RWMutex
	RequestCount int64         `json:"request_count"`
	CacheHits    int64         `json:"cache_hits"`
	CacheMisses  int64         `json:"cache_misses"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

// bookService implements BookService interface with enhanced features
type bookService struct {
	bookRepo  repository.BookRepository
	cache     *cache.BookCache
	processor *workers.BookProcessor
	metrics   *ServiceMetrics
}

// NewBookService creates a new enhanced book service
func NewBookService(bookRepo repository.BookRepository, cache *cache.BookCache, processor *workers.BookProcessor) BookService {
	return &bookService{
		bookRepo:  bookRepo,
		cache:     cache,
		processor: processor,
		metrics:   &ServiceMetrics{},
	}
}

// CreateBook creates a new book with enhanced concurrent processing
func (s *bookService) CreateBook(req *models.CreateBookRequest) (*models.Book, error) {
	start := time.Now()
	defer s.recordMetrics(start)

	// Use channels for validation pipeline
	validationChan := make(chan error, 3)

	// Concurrent validation checks
	go s.validateTitle(req.Title, validationChan)
	go s.validateAuthor(req.Author, validationChan)
	go s.validateISBN(req.ISBN, validationChan)

	// Collect validation results
	for i := 0; i < 3; i++ {
		if err := <-validationChan; err != nil {
			return nil, err
		}
	}

	// Check ISBN uniqueness concurrently
	existsChan := make(chan bool, 1)
	errChan := make(chan error, 1)

	go func() {
		exists, err := s.bookRepo.ExistsByISBN(req.ISBN, nil)
		if err != nil {
			errChan <- err
			return
		}
		existsChan <- exists
	}()

	select {
	case exists := <-existsChan:
		if exists {
			return nil, fmt.Errorf("book with ISBN %s already exists", req.ISBN)
		}
	case err := <-errChan:
		return nil, fmt.Errorf("failed to check ISBN uniqueness: %w", err)
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("validation timeout")
	}

	// Normalize data
	s.normalizeBookData(req)

	// Create the book
	book, err := s.bookRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	// Cache the new book
	if s.cache != nil {
		s.cache.SetBook(book)
		// Invalidate book list caches since we added a new book
		s.cache.Clear() // Clear all cached book lists
	}

	// Submit background job for post-processing
	if s.processor != nil {
		job := workers.BookJob{
			ID:       uuid.New().String(),
			Type:     workers.JobTypeNotify,
			BookData: req,
		}
		s.processor.SubmitJob(job)
	}

	return book, nil
}

// GetBookByID retrieves a book with Redis caching
func (s *bookService) GetBookByID(id uuid.UUID) (*models.Book, error) {
	start := time.Now()
	defer s.recordMetrics(start)

	// Try cache first (Redis + in-memory fallback)
	if s.cache != nil {
		if book, found := s.cache.GetBook(id); found {
			s.metrics.mu.Lock()
			s.metrics.CacheHits++
			s.metrics.mu.Unlock()
			return book, nil
		}

		// Cache miss
		s.metrics.mu.Lock()
		s.metrics.CacheMisses++
		s.metrics.mu.Unlock()
	}

	book, err := s.bookRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	// Cache the result in both Redis and in-memory
	if s.cache != nil {
		s.cache.SetBook(book)
	}

	return book, nil
}

// GetAllBooks retrieves books with Redis caching for enhanced performance
func (s *bookService) GetAllBooks(filter models.BookFilter) (*models.BooksListResponse, error) {
	start := time.Now()
	defer s.recordMetrics(start)

	// Try cache first (Redis + in-memory fallback)
	if s.cache != nil {
		if response, found := s.cache.GetBookList(filter); found {
			s.metrics.mu.Lock()
			s.metrics.CacheHits++
			s.metrics.mu.Unlock()
			return response, nil
		}

		// Cache miss
		s.metrics.mu.Lock()
		s.metrics.CacheMisses++
		s.metrics.mu.Unlock()
	}

	// Set reasonable defaults
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	books, total, err := s.bookRepo.GetAll(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}

	response := &models.BooksListResponse{
		Books:  books,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	// Cache the results in both Redis and in-memory
	if s.cache != nil {
		s.cache.SetBookList(filter, response)
	}

	return response, nil
}

// UpdateBook updates a book with enhanced validation and caching
func (s *bookService) UpdateBook(id uuid.UUID, req *models.UpdateBookRequest) (*models.Book, error) {
	start := time.Now()
	defer s.recordMetrics(start)

	// Check if book exists
	_, err := s.bookRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	// Validate business rules
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Check ISBN uniqueness if being updated
	if req.ISBN != nil {
		exists, err := s.bookRepo.ExistsByISBN(*req.ISBN, &id)
		if err != nil {
			return nil, fmt.Errorf("failed to check ISBN uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("book with ISBN %s already exists", *req.ISBN)
		}
		// Normalize ISBN
		normalizedISBN := normalizeISBN(*req.ISBN)
		req.ISBN = &normalizedISBN
	}

	// Normalize other fields
	if req.Title != nil {
		normalized := strings.TrimSpace(*req.Title)
		req.Title = &normalized
	}
	if req.Author != nil {
		normalized := strings.TrimSpace(*req.Author)
		req.Author = &normalized
	}
	if req.Publisher != nil {
		normalized := strings.TrimSpace(*req.Publisher)
		req.Publisher = &normalized
	}
	if req.Genre != nil {
		normalized := strings.TrimSpace(*req.Genre)
		req.Genre = &normalized
	}
	if req.Language != nil {
		normalized := strings.TrimSpace(*req.Language)
		req.Language = &normalized
	}

	// Update the book
	book, err := s.bookRepo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update book: %w", err)
	}

	// Update cache
	if s.cache != nil {
		s.cache.SetBook(book)
		// Invalidate related cache entries (also invalidates book lists)
		s.cache.InvalidateBook(book.ID)
	}

	return book, nil
}

// DeleteBook deletes a book with cache invalidation
func (s *bookService) DeleteBook(id uuid.UUID) error {
	start := time.Now()
	defer s.recordMetrics(start)

	// Check if book exists
	_, err := s.bookRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("book not found: %w", err)
	}

	if err := s.bookRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	// Remove from cache
	if s.cache != nil {
		// InvalidateBook will handle both individual book and book list invalidation
		s.cache.InvalidateBook(id)
	}

	return nil
}

// BulkCreateBooks demonstrates concurrent bulk operations
func (s *bookService) BulkCreateBooks(requests []*models.CreateBookRequest) ([]*models.Book, []error) {
	if len(requests) == 0 {
		return nil, nil
	}

	start := time.Now()
	defer s.recordMetrics(start)

	// Channel to control concurrency
	semaphore := make(chan struct{}, 10) // Max 10 concurrent operations

	// Results channels
	resultChan := make(chan struct {
		index int
		book  *models.Book
		err   error
	}, len(requests))

	// Process each request concurrently
	var wg sync.WaitGroup
	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *models.CreateBookRequest) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			book, err := s.CreateBook(request)
			resultChan <- struct {
				index int
				book  *models.Book
				err   error
			}{index, book, err}
		}(i, req)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	books := make([]*models.Book, len(requests))
	errors := make([]error, len(requests))

	for result := range resultChan {
		books[result.index] = result.book
		errors[result.index] = result.err
	}

	return books, errors
}

// GetMetrics returns service metrics safely (without copying mutex)
func (s *bookService) GetMetrics() ServiceMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Return a copy without the mutex
	return ServiceMetrics{
		RequestCount: s.metrics.RequestCount,
		CacheHits:    s.metrics.CacheHits,
		CacheMisses:  s.metrics.CacheMisses,
		AvgLatency:   s.metrics.AvgLatency,
	}
}

// Shutdown gracefully shuts down the service
func (s *bookService) Shutdown(ctx context.Context) error {
	// Stop background processor
	if s.processor != nil {
		s.processor.Stop()
	}

	// Stop cache cleanup
	if s.cache != nil {
		s.cache.Shutdown()
	}

	return nil
}

// Helper methods for validation
func (s *bookService) validateTitle(title string, errChan chan<- error) {
	if strings.TrimSpace(title) == "" {
		errChan <- fmt.Errorf("title is required")
		return
	}
	errChan <- nil
}

func (s *bookService) validateAuthor(author string, errChan chan<- error) {
	if strings.TrimSpace(author) == "" {
		errChan <- fmt.Errorf("author is required")
		return
	}
	errChan <- nil
}

func (s *bookService) validateISBN(isbn string, errChan chan<- error) {
	if strings.TrimSpace(isbn) == "" {
		errChan <- fmt.Errorf("ISBN is required")
		return
	}

	// Add ISBN format validation
	normalized := normalizeISBN(isbn)
	if !isValidISBN(normalized) {
		errChan <- fmt.Errorf("invalid ISBN format")
		return
	}

	errChan <- nil
}

func (s *bookService) normalizeBookData(req *models.CreateBookRequest) {
	req.Title = strings.TrimSpace(req.Title)
	req.Author = strings.TrimSpace(req.Author)
	req.ISBN = normalizeISBN(req.ISBN)
	req.Publisher = strings.TrimSpace(req.Publisher)
	req.Genre = strings.TrimSpace(req.Genre)
	req.Language = strings.TrimSpace(req.Language)
}

func (s *bookService) generateCacheKey(filter models.BookFilter) string {
	return fmt.Sprintf("books:%s_%s_%s_%v_%d_%d",
		filter.Author, filter.Genre, filter.Language,
		filter.Available, filter.Limit, filter.Offset)
}

func (s *bookService) recordMetrics(start time.Time) {
	duration := time.Since(start)

	s.metrics.mu.Lock()
	s.metrics.RequestCount++
	// Simple moving average
	s.metrics.AvgLatency = (s.metrics.AvgLatency + duration) / 2
	s.metrics.mu.Unlock()
}

// validateCreateRequest validates the create book request
func (s *bookService) validateCreateRequest(req *models.CreateBookRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(req.Author) == "" {
		return fmt.Errorf("author is required")
	}
	if strings.TrimSpace(req.ISBN) == "" {
		return fmt.Errorf("ISBN is required")
	}
	if !isValidISBN(req.ISBN) {
		return fmt.Errorf("invalid ISBN format")
	}
	if req.Pages <= 0 {
		return fmt.Errorf("pages must be greater than 0")
	}
	return nil
}

// validateUpdateRequest validates the update book request
func (s *bookService) validateUpdateRequest(req *models.UpdateBookRequest) error {
	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if req.Author != nil && strings.TrimSpace(*req.Author) == "" {
		return fmt.Errorf("author cannot be empty")
	}
	if req.ISBN != nil {
		if strings.TrimSpace(*req.ISBN) == "" {
			return fmt.Errorf("ISBN cannot be empty")
		}
		if !isValidISBN(*req.ISBN) {
			return fmt.Errorf("invalid ISBN format")
		}
	}
	if req.Pages != nil && *req.Pages <= 0 {
		return fmt.Errorf("pages must be greater than 0")
	}
	return nil
}

// normalizeISBN removes hyphens and spaces from ISBN
func normalizeISBN(isbn string) string {
	return strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")
}

// isValidISBN checks if the ISBN format is valid (basic validation)
func isValidISBN(isbn string) bool {
	normalized := normalizeISBN(isbn)
	// Basic validation: ISBN-10 (10 digits) or ISBN-13 (13 digits)
	if len(normalized) != 10 && len(normalized) != 13 {
		return false
	}
	// Additional validation logic can be added here (checksum validation)
	return true
}
