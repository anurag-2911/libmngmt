package handlers

import (
	"context"
	"encoding/json"
	"libmngmt/internal/errors"
	"libmngmt/internal/middleware"
	"libmngmt/internal/models"
	"libmngmt/internal/service"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// BookHandler handles HTTP requests for books with enhanced features
type BookHandler struct {
	bookService    service.BookService
	requestLimiter chan struct{}
	activeRequests sync.WaitGroup
	metrics        *HandlerMetrics
}

// HandlerMetrics tracks handler performance
type HandlerMetrics struct {
	mu              sync.RWMutex
	totalRequests   int64
	activeRequests  int64
	requestDuration map[string]time.Duration
}

// NewBookHandler creates a new enhanced book handler
func NewBookHandler(bookService service.BookService) *BookHandler {
	return &BookHandler{
		bookService:    bookService,
		requestLimiter: make(chan struct{}, 100), // Limit to 100 concurrent requests
		metrics: &HandlerMetrics{
			requestDuration: make(map[string]time.Duration),
		},
	}
}

// CreateBook handles POST /api/books with concurrent processing
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics("CreateBook", start)

	// Rate limiting
	select {
	case h.requestLimiter <- struct{}{}:
		defer func() { <-h.requestLimiter }()
	default:
		h.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded", "Too many concurrent requests")
		return
	}

	h.activeRequests.Add(1)
	defer h.activeRequests.Done()

	// Context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Parse request in goroutine
	requestChan := make(chan *models.CreateBookRequest, 1)
	errChan := make(chan error, 1)

	go func() {
		var req models.CreateBookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errChan <- err
			return
		}
		requestChan <- &req
	}()

	var req *models.CreateBookRequest
	select {
	case req = <-requestChan:
	case err := <-errChan:
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	case <-ctx.Done():
		h.writeErrorResponse(w, http.StatusRequestTimeout, "Request timeout", "Request parsing timed out")
		return
	}

	// Create book using service
	bookChan := make(chan *models.Book, 1)
	errorChan := make(chan error, 1)

	go func() {
		book, err := h.bookService.CreateBook(req)
		if err != nil {
			errorChan <- err
			return
		}
		bookChan <- book
	}()

	select {
	case book := <-bookChan:
		h.writeSuccessResponse(w, http.StatusCreated, "Book created successfully", book)
	case err := <-errorChan:
		if isValidationError(err) {
			h.writeErrorResponse(w, http.StatusBadRequest, "Validation error", err.Error())
		} else if isDuplicateError(err) {
			h.writeErrorResponse(w, http.StatusConflict, "Duplicate resource", err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", err.Error())
		}
	case <-ctx.Done():
		h.writeErrorResponse(w, http.StatusRequestTimeout, "Request timeout", "Book creation timed out")
	}
}

// GetBook handles GET /api/books/{id} with caching
func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics("GetBook", start)

	h.activeRequests.Add(1)
	defer h.activeRequests.Done()

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "ID must be a valid UUID")
		return
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get book using service with caching
	bookChan := make(chan *models.Book, 1)
	errChan := make(chan error, 1)

	go func() {
		book, err := h.bookService.GetBookByID(id)
		if err != nil {
			errChan <- err
			return
		}
		bookChan <- book
	}()

	select {
	case book := <-bookChan:
		h.writeSuccessResponse(w, http.StatusOK, "Book retrieved successfully", book)
	case err := <-errChan:
		if isNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, "Book not found", err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", err.Error())
		}
	case <-ctx.Done():
		h.writeErrorResponse(w, http.StatusRequestTimeout, "Request timeout", "Book retrieval timed out")
	}
}

// GetBooks handles GET /api/books with concurrent processing
func (h *BookHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics("GetBooks", start)

	h.activeRequests.Add(1)
	defer h.activeRequests.Done()

	// Parse query parameters concurrently
	filterChan := make(chan models.BookFilter, 1)
	go func() {
		filter := models.BookFilter{}

		if author := r.URL.Query().Get("author"); author != "" {
			filter.Author = author
		}
		if genre := r.URL.Query().Get("genre"); genre != "" {
			filter.Genre = genre
		}
		if language := r.URL.Query().Get("language"); language != "" {
			filter.Language = language
		}
		if availableStr := r.URL.Query().Get("available"); availableStr != "" {
			if available, err := strconv.ParseBool(availableStr); err == nil {
				filter.Available = &available
			}
		}
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				filter.Limit = limit
			}
		}
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
				filter.Offset = offset
			}
		}

		filterChan <- filter
	}()

	filter := <-filterChan

	// Context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Get books using service
	responseChan := make(chan *models.BooksListResponse, 1)
	errChan := make(chan error, 1)

	go func() {
		response, err := h.bookService.GetAllBooks(filter)
		if err != nil {
			errChan <- err
			return
		}
		responseChan <- response
	}()

	select {
	case response := <-responseChan:
		h.writeSuccessResponse(w, http.StatusOK, "Books retrieved successfully", response)
	case err := <-errChan:
		h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", err.Error())
	case <-ctx.Done():
		h.writeErrorResponse(w, http.StatusRequestTimeout, "Request timeout", "Books retrieval timed out")
	}
}

// UpdateBook handles PUT /api/books/{id}
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "ID must be a valid UUID")
		return
	}

	var req models.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	book, err := h.bookService.UpdateBook(id, &req)
	if err != nil {
		if isNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, "Book not found", err.Error())
			return
		}
		if isValidationError(err) {
			h.writeErrorResponse(w, http.StatusBadRequest, "Validation error", err.Error())
			return
		}
		if isDuplicateError(err) {
			h.writeErrorResponse(w, http.StatusConflict, "Duplicate resource", err.Error())
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "Book updated successfully", book)
}

// DeleteBook handles DELETE /api/books/{id}
func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "ID must be a valid UUID")
		return
	}

	err = h.bookService.DeleteBook(id)
	if err != nil {
		if isNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, "Book not found", err.Error())
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "Book deleted successfully", nil)
}

// BulkCreateBooks handles bulk book creation
func (h *BookHandler) BulkCreateBooks(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics("BulkCreateBooks", start)

	h.activeRequests.Add(1)
	defer h.activeRequests.Done()

	// Context with extended timeout for bulk operations
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Parse bulk request
	var requests []*models.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	if len(requests) == 0 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Empty request", "No books provided")
		return
	}

	if len(requests) > 100 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Too many books", "Maximum 100 books per request")
		return
	}

	// Process bulk creation
	resultChan := make(chan struct {
		books  []*models.Book
		errors []error
	}, 1)

	go func() {
		books, errors := h.bookService.BulkCreateBooks(requests)
		resultChan <- struct {
			books  []*models.Book
			errors []error
		}{books, errors}
	}()

	select {
	case result := <-resultChan:
		// Prepare response
		successCount := 0
		var successBooks []*models.Book
		var errorDetails []map[string]interface{}

		for i, book := range result.books {
			if result.errors[i] == nil {
				successCount++
				successBooks = append(successBooks, book)
			} else {
				errorDetails = append(errorDetails, map[string]interface{}{
					"index": i,
					"error": result.errors[i].Error(),
					"book":  requests[i],
				})
			}
		}

		response := map[string]interface{}{
			"total_requested": len(requests),
			"successful":      successCount,
			"failed":          len(requests) - successCount,
			"books":           successBooks,
			"errors":          errorDetails,
		}

		if successCount == len(requests) {
			h.writeSuccessResponse(w, http.StatusCreated, "All books created successfully", response)
		} else if successCount > 0 {
			h.writeSuccessResponse(w, http.StatusPartialContent, "Some books created successfully", response)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
		}

	case <-ctx.Done():
		h.writeErrorResponse(w, http.StatusRequestTimeout, "Request timeout", "Bulk creation timed out")
	}
}

// GetMetrics returns handler metrics
func (h *BookHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.metrics.mu.RLock()
	metrics := map[string]interface{}{
		"total_requests":   h.metrics.totalRequests,
		"active_requests":  h.metrics.activeRequests,
		"request_duration": h.metrics.requestDuration,
	}
	h.metrics.mu.RUnlock()

	// Add service metrics safely
	serviceMetrics := h.bookService.GetMetrics()
	metrics["service_metrics"] = map[string]interface{}{
		"request_count": serviceMetrics.RequestCount,
		"cache_hits":    serviceMetrics.CacheHits,
		"cache_misses":  serviceMetrics.CacheMisses,
		"avg_latency":   serviceMetrics.AvgLatency,
	}

	h.writeSuccessResponse(w, http.StatusOK, "Metrics retrieved successfully", metrics)
}

// Graceful shutdown
func (h *BookHandler) Shutdown(ctx context.Context) error {
	// Wait for active requests to complete
	done := make(chan struct{})
	go func() {
		h.activeRequests.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// HealthCheck handles GET /health
func (h *BookHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "library-management-api",
		"timestamp": "2025-01-23T10:00:00Z",
	}
	h.writeSuccessResponse(w, http.StatusOK, "Service is healthy", response)
}

// Helper methods
func (h *BookHandler) recordMetrics(operation string, start time.Time) {
	duration := time.Since(start)

	h.metrics.mu.Lock()
	h.metrics.totalRequests++
	h.metrics.requestDuration[operation] = duration
	h.metrics.mu.Unlock()
}

func (h *BookHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.WriteHeader(statusCode)
	response := models.SuccessResponse{
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *BookHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	w.WriteHeader(statusCode)
	response := models.ErrorResponse{
		Error:   error,
		Message: message,
		Code:    statusCode,
	}
	json.NewEncoder(w).Encode(response)
}

// writeStructuredErrorResponse uses the new error handling system
func (h *BookHandler) writeStructuredErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, error, message string) {
	// Get request ID from context
	requestID := middleware.GetRequestID(r.Context())

	// Create appropriate error based on status code
	var appErr *errors.AppError
	switch statusCode {
	case http.StatusBadRequest:
		appErr = errors.Validation(error, message)
	case http.StatusNotFound:
		appErr = errors.NotFound(error)
	case http.StatusConflict:
		appErr = errors.Conflict(error, message)
	case http.StatusRequestTimeout:
		appErr = errors.New(errors.CodeTimeout, error, message)
	case http.StatusTooManyRequests:
		appErr = errors.New(errors.CodeRateLimit, error, message)
	default:
		appErr = errors.Internal(error, message)
	}

	// Write structured error response
	errors.WriteErrorResponse(w, appErr, requestID)
}

// Error classification helpers
func isValidationError(err error) bool {
	errMsg := err.Error()
	validationKeywords := []string{
		"is required",
		"cannot be empty",
		"invalid",
		"must be greater than",
		"format",
	}
	for _, keyword := range validationKeywords {
		if contains(errMsg, keyword) {
			return true
		}
	}
	return false
}

func isDuplicateError(err error) bool {
	return contains(err.Error(), "already exists")
}

func isNotFoundError(err error) bool {
	return contains(err.Error(), "not found")
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					containsSubstring(str, substr))))
}

func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
