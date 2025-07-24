package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"libmngmt/internal/models"
	"libmngmt/internal/service"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBookService is a mock implementation of BookService for testing
type MockBookService struct {
	mock.Mock
}

func (m *MockBookService) CreateBook(req *models.CreateBookRequest) (*models.Book, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) GetBookByID(id uuid.UUID) (*models.Book, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) GetAllBooks(filter models.BookFilter) (*models.BooksListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BooksListResponse), args.Error(1)
}

func (m *MockBookService) UpdateBook(id uuid.UUID, req *models.UpdateBookRequest) (*models.Book, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) DeleteBook(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBookService) BulkCreateBooks(reqs []*models.CreateBookRequest) ([]*models.Book, []error) {
	args := m.Called(reqs)
	if args.Get(0) == nil {
		return nil, args.Get(1).([]error)
	}
	return args.Get(0).([]*models.Book), args.Get(1).([]error)
}

func (m *MockBookService) GetMetrics() service.ServiceMetrics {
	args := m.Called()
	return args.Get(0).(service.ServiceMetrics)
}

func (m *MockBookService) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test utilities
func createTestBook() *models.Book {
	id := uuid.New()
	now := time.Now()
	return &models.Book{
		ID:          id,
		Title:       "Test Book",
		Author:      "Test Author",
		ISBN:        "9781234567890",
		Publisher:   "Test Publisher",
		Genre:       "Fiction",
		PublishedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Pages:       300,
		Language:    "English",
		Available:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createValidCreateRequest() *models.CreateBookRequest {
	return &models.CreateBookRequest{
		Title:       "Test Book",
		Author:      "Test Author",
		ISBN:        "9781234567890",
		Publisher:   "Test Publisher",
		Genre:       "Fiction",
		PublishedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Pages:       300,
		Language:    "English",
	}
}

func setupHandlerTest() (*BookHandler, *MockBookService) {
	mockService := &MockBookService{}
	handler := NewBookHandler(mockService)
	return handler, mockService
}

// Test NewBookHandler
func TestNewBookHandler(t *testing.T) {
	t.Run("create new book handler", func(t *testing.T) {
		mockService := &MockBookService{}
		handler := NewBookHandler(mockService)

		assert.NotNil(t, handler)
		assert.Equal(t, mockService, handler.bookService)
		assert.NotNil(t, handler.requestLimiter)
		assert.NotNil(t, handler.metrics)
		assert.Equal(t, 100, cap(handler.requestLimiter))
	})
}

// Test CreateBook handler
func TestBookHandler_CreateBook(t *testing.T) {
	t.Run("create book successfully", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		req := createValidCreateRequest()
		book := createTestBook()

		mockService.On("CreateBook", req).Return(book, nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/books", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBook(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Book created successfully", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("create book with invalid JSON", func(t *testing.T) {
		handler, _ := setupHandlerTest()

		httpReq := httptest.NewRequest("POST", "/api/books", strings.NewReader("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBook(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid JSON", response["error"])
		assert.Contains(t, response["message"].(string), "invalid character")
	})

	t.Run("create book with service error", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		req := createValidCreateRequest()

		mockService.On("CreateBook", req).Return(nil, errors.New("service error"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/books", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBook(w, httpReq)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])
		assert.Equal(t, "service error", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("create book with validation error", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		req := createValidCreateRequest()

		mockService.On("CreateBook", req).Return(nil, errors.New("title is required"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/books", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBook(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Validation error", response["error"])
		assert.Equal(t, "title is required", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("create book with duplicate ISBN error", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		req := createValidCreateRequest()

		mockService.On("CreateBook", req).Return(nil, errors.New("book with ISBN 9781234567890 already exists"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/books", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBook(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Duplicate resource", response["error"])
		assert.Contains(t, response["message"].(string), "book with ISBN 9781234567890 already exists")

		mockService.AssertExpectations(t)
	})

	t.Run("create book with request timeout", func(t *testing.T) {
		handler, _ := setupHandlerTest()

		// Fill up the rate limiter to simulate congestion
		for i := 0; i < 100; i++ {
			handler.requestLimiter <- struct{}{}
		}

		req := createValidCreateRequest()
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/books", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBook(w, httpReq)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Rate limit exceeded", response["error"])
		assert.Equal(t, "Too many concurrent requests", response["message"])
	})
}

// Test GetBook handler
func TestBookHandler_GetBook(t *testing.T) {
	t.Run("get book successfully", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		book := createTestBook()

		mockService.On("GetBookByID", book.ID).Return(book, nil)

		httpReq := httptest.NewRequest("GET", "/api/books/"+book.ID.String(), nil)
		httpReq = mux.SetURLVars(httpReq, map[string]string{"id": book.ID.String()})
		w := httptest.NewRecorder()

		handler.GetBook(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Book retrieved successfully", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("get book with invalid UUID", func(t *testing.T) {
		handler, _ := setupHandlerTest()

		httpReq := httptest.NewRequest("GET", "/api/books/invalid-uuid", nil)
		httpReq = mux.SetURLVars(httpReq, map[string]string{"id": "invalid-uuid"})
		w := httptest.NewRecorder()

		handler.GetBook(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid book ID", response["error"])
		assert.Equal(t, "ID must be a valid UUID", response["message"])
	})

	t.Run("get book not found", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		id := uuid.New()

		mockService.On("GetBookByID", id).Return(nil, errors.New("book not found"))

		httpReq := httptest.NewRequest("GET", "/api/books/"+id.String(), nil)
		httpReq = mux.SetURLVars(httpReq, map[string]string{"id": id.String()})
		w := httptest.NewRecorder()

		handler.GetBook(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"].(string), "not found")

		mockService.AssertExpectations(t)
	})

	t.Run("get book with service error", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		id := uuid.New()

		mockService.On("GetBookByID", id).Return(nil, errors.New("service error"))

		httpReq := httptest.NewRequest("GET", "/api/books/"+id.String(), nil)
		httpReq = mux.SetURLVars(httpReq, map[string]string{"id": id.String()})
		w := httptest.NewRecorder()

		handler.GetBook(w, httpReq)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])
		assert.Equal(t, "service error", response["message"])

		mockService.AssertExpectations(t)
	})
}

// Test GetBooks handler
func TestBookHandler_GetBooks(t *testing.T) {
	t.Run("get books successfully", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		books := []models.Book{*createTestBook(), *createTestBook()}
		filter := models.BookFilter{Limit: 0, Offset: 0}

		mockService.On("GetAllBooks", filter).Return(&models.BooksListResponse{
			Books:  books,
			Total:  2,
			Limit:  10,
			Offset: 0,
		}, nil)

		httpReq := httptest.NewRequest("GET", "/api/books", nil)
		w := httptest.NewRecorder()

		handler.GetBooks(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Books retrieved successfully", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("get books with filters", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		available := true
		books := []models.Book{*createTestBook()}
		filter := models.BookFilter{
			Author:    "tolkien",
			Genre:     "fantasy",
			Language:  "English",
			Available: &available,
			Limit:     5,
			Offset:    10,
		}

		mockService.On("GetAllBooks", filter).Return(&models.BooksListResponse{
			Books:  books,
			Total:  1,
			Limit:  5,
			Offset: 10,
		}, nil)

		params := url.Values{}
		params.Add("author", "tolkien")
		params.Add("genre", "fantasy")
		params.Add("language", "English")
		params.Add("available", "true")
		params.Add("limit", "5")
		params.Add("offset", "10")

		httpReq := httptest.NewRequest("GET", "/api/books?"+params.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetBooks(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("get books with service error", func(t *testing.T) {
		handler, mockService := setupHandlerTest()

		filter := models.BookFilter{Limit: 0, Offset: 0}

		mockService.On("GetAllBooks", filter).Return(nil, errors.New("service error"))

		httpReq := httptest.NewRequest("GET", "/api/books", nil)
		w := httptest.NewRecorder()

		handler.GetBooks(w, httpReq)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])
		assert.Equal(t, "service error", response["message"])

		mockService.AssertExpectations(t)
	})
}
