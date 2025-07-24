package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBook_Creation(t *testing.T) {
	t.Run("create book with all fields", func(t *testing.T) {
		id := uuid.New()
		now := time.Now()
		publishedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		book := Book{
			ID:          id,
			Title:       "Test Book",
			Author:      "Test Author",
			ISBN:        "9781234567890",
			Publisher:   "Test Publisher",
			Genre:       "Fiction",
			PublishedAt: publishedAt,
			Pages:       300,
			Language:    "English",
			Available:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		assert.Equal(t, id, book.ID)
		assert.Equal(t, "Test Book", book.Title)
		assert.Equal(t, "Test Author", book.Author)
		assert.Equal(t, "9781234567890", book.ISBN)
		assert.Equal(t, "Test Publisher", book.Publisher)
		assert.Equal(t, "Fiction", book.Genre)
		assert.Equal(t, publishedAt, book.PublishedAt)
		assert.Equal(t, 300, book.Pages)
		assert.Equal(t, "English", book.Language)
		assert.True(t, book.Available)
		assert.Equal(t, now, book.CreatedAt)
		assert.Equal(t, now, book.UpdatedAt)
	})

	t.Run("create book with minimal fields", func(t *testing.T) {
		book := Book{
			Title:  "Minimal Book",
			Author: "Minimal Author",
			ISBN:   "1234567890",
		}

		assert.Equal(t, "Minimal Book", book.Title)
		assert.Equal(t, "Minimal Author", book.Author)
		assert.Equal(t, "1234567890", book.ISBN)
		assert.Empty(t, book.Publisher)
		assert.Empty(t, book.Genre)
		assert.Zero(t, book.Pages)
		assert.Empty(t, book.Language)
		assert.False(t, book.Available) // default value
	})
}

func TestCreateBookRequest_Creation(t *testing.T) {
	t.Run("create valid book request", func(t *testing.T) {
		publishedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		req := CreateBookRequest{
			Title:       "Test Book",
			Author:      "Test Author",
			ISBN:        "9781234567890",
			Publisher:   "Test Publisher",
			Genre:       "Fiction",
			PublishedAt: publishedAt,
			Pages:       300,
			Language:    "English",
		}

		assert.Equal(t, "Test Book", req.Title)
		assert.Equal(t, "Test Author", req.Author)
		assert.Equal(t, "9781234567890", req.ISBN)
		assert.Equal(t, "Test Publisher", req.Publisher)
		assert.Equal(t, "Fiction", req.Genre)
		assert.Equal(t, publishedAt, req.PublishedAt)
		assert.Equal(t, 300, req.Pages)
		assert.Equal(t, "English", req.Language)
	})

	t.Run("create minimal book request", func(t *testing.T) {
		req := CreateBookRequest{
			Title:  "Minimal Book",
			Author: "Minimal Author",
			ISBN:   "1234567890",
			Pages:  100,
		}

		assert.Equal(t, "Minimal Book", req.Title)
		assert.Equal(t, "Minimal Author", req.Author)
		assert.Equal(t, "1234567890", req.ISBN)
		assert.Equal(t, 100, req.Pages)
		assert.Empty(t, req.Publisher)
		assert.Empty(t, req.Genre)
		assert.Empty(t, req.Language)
		assert.Zero(t, req.PublishedAt)
	})
}

func TestUpdateBookRequest_Creation(t *testing.T) {
	t.Run("create update request with all fields", func(t *testing.T) {
		title := "Updated Title"
		author := "Updated Author"
		isbn := "9780987654321"
		publisher := "Updated Publisher"
		genre := "Updated Genre"
		publishedAt := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
		pages := 250
		language := "Spanish"
		available := false

		req := UpdateBookRequest{
			Title:       &title,
			Author:      &author,
			ISBN:        &isbn,
			Publisher:   &publisher,
			Genre:       &genre,
			PublishedAt: &publishedAt,
			Pages:       &pages,
			Language:    &language,
			Available:   &available,
		}

		assert.NotNil(t, req.Title)
		assert.Equal(t, "Updated Title", *req.Title)
		assert.NotNil(t, req.Author)
		assert.Equal(t, "Updated Author", *req.Author)
		assert.NotNil(t, req.ISBN)
		assert.Equal(t, "9780987654321", *req.ISBN)
		assert.NotNil(t, req.Publisher)
		assert.Equal(t, "Updated Publisher", *req.Publisher)
		assert.NotNil(t, req.Genre)
		assert.Equal(t, "Updated Genre", *req.Genre)
		assert.NotNil(t, req.PublishedAt)
		assert.Equal(t, publishedAt, *req.PublishedAt)
		assert.NotNil(t, req.Pages)
		assert.Equal(t, 250, *req.Pages)
		assert.NotNil(t, req.Language)
		assert.Equal(t, "Spanish", *req.Language)
		assert.NotNil(t, req.Available)
		assert.False(t, *req.Available)
	})

	t.Run("create partial update request", func(t *testing.T) {
		title := "Only Title Updated"
		available := true

		req := UpdateBookRequest{
			Title:     &title,
			Available: &available,
		}

		assert.NotNil(t, req.Title)
		assert.Equal(t, "Only Title Updated", *req.Title)
		assert.NotNil(t, req.Available)
		assert.True(t, *req.Available)
		assert.Nil(t, req.Author)
		assert.Nil(t, req.ISBN)
		assert.Nil(t, req.Publisher)
		assert.Nil(t, req.Genre)
		assert.Nil(t, req.PublishedAt)
		assert.Nil(t, req.Pages)
		assert.Nil(t, req.Language)
	})

	t.Run("create empty update request", func(t *testing.T) {
		req := UpdateBookRequest{}

		assert.Nil(t, req.Title)
		assert.Nil(t, req.Author)
		assert.Nil(t, req.ISBN)
		assert.Nil(t, req.Publisher)
		assert.Nil(t, req.Genre)
		assert.Nil(t, req.PublishedAt)
		assert.Nil(t, req.Pages)
		assert.Nil(t, req.Language)
		assert.Nil(t, req.Available)
	})
}

func TestBookFilter_Creation(t *testing.T) {
	t.Run("create filter with all fields", func(t *testing.T) {
		available := true

		filter := BookFilter{
			Author:    "Tolkien",
			Genre:     "Fantasy",
			Language:  "English",
			Available: &available,
			Limit:     20,
			Offset:    10,
		}

		assert.Equal(t, "Tolkien", filter.Author)
		assert.Equal(t, "Fantasy", filter.Genre)
		assert.Equal(t, "English", filter.Language)
		assert.NotNil(t, filter.Available)
		assert.True(t, *filter.Available)
		assert.Equal(t, 20, filter.Limit)
		assert.Equal(t, 10, filter.Offset)
	})

	t.Run("create empty filter", func(t *testing.T) {
		filter := BookFilter{}

		assert.Empty(t, filter.Author)
		assert.Empty(t, filter.Genre)
		assert.Empty(t, filter.Language)
		assert.Nil(t, filter.Available)
		assert.Zero(t, filter.Limit)
		assert.Zero(t, filter.Offset)
	})

	t.Run("create filter with available false", func(t *testing.T) {
		available := false

		filter := BookFilter{
			Available: &available,
		}

		assert.NotNil(t, filter.Available)
		assert.False(t, *filter.Available)
	})
}

func TestErrorResponse_Creation(t *testing.T) {
	t.Run("create error response", func(t *testing.T) {
		err := ErrorResponse{
			Error:   "Validation Error",
			Message: "Title is required",
			Code:    400,
		}

		assert.Equal(t, "Validation Error", err.Error)
		assert.Equal(t, "Title is required", err.Message)
		assert.Equal(t, 400, err.Code)
	})

	t.Run("create error response without message", func(t *testing.T) {
		err := ErrorResponse{
			Error: "Internal Server Error",
			Code:  500,
		}

		assert.Equal(t, "Internal Server Error", err.Error)
		assert.Empty(t, err.Message)
		assert.Equal(t, 500, err.Code)
	})
}

func TestSuccessResponse_Creation(t *testing.T) {
	t.Run("create success response with book data", func(t *testing.T) {
		book := Book{
			ID:     uuid.New(),
			Title:  "Test Book",
			Author: "Test Author",
		}

		resp := SuccessResponse{
			Message: "Book created successfully",
			Data:    book,
		}

		assert.Equal(t, "Book created successfully", resp.Message)
		assert.Equal(t, book, resp.Data)
	})

	t.Run("create success response without data", func(t *testing.T) {
		resp := SuccessResponse{
			Message: "Operation completed successfully",
		}

		assert.Equal(t, "Operation completed successfully", resp.Message)
		assert.Nil(t, resp.Data)
	})

	t.Run("create success response with string data", func(t *testing.T) {
		resp := SuccessResponse{
			Message: "Status retrieved",
			Data:    "healthy",
		}

		assert.Equal(t, "Status retrieved", resp.Message)
		assert.Equal(t, "healthy", resp.Data)
	})
}

func TestBooksListResponse_Creation(t *testing.T) {
	t.Run("create books list response", func(t *testing.T) {
		books := []Book{
			{ID: uuid.New(), Title: "Book 1", Author: "Author 1"},
			{ID: uuid.New(), Title: "Book 2", Author: "Author 2"},
		}

		resp := BooksListResponse{
			Books:  books,
			Total:  2,
			Limit:  10,
			Offset: 0,
		}

		assert.Equal(t, books, resp.Books)
		assert.Len(t, resp.Books, 2)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, 10, resp.Limit)
		assert.Equal(t, 0, resp.Offset)
	})

	t.Run("create empty books list response", func(t *testing.T) {
		resp := BooksListResponse{
			Books:  []Book{},
			Total:  0,
			Limit:  50,
			Offset: 0,
		}

		assert.Empty(t, resp.Books)
		assert.Equal(t, 0, resp.Total)
		assert.Equal(t, 50, resp.Limit)
		assert.Equal(t, 0, resp.Offset)
	})

	t.Run("create paginated books list response", func(t *testing.T) {
		books := []Book{
			{ID: uuid.New(), Title: "Book 21", Author: "Author 21"},
			{ID: uuid.New(), Title: "Book 22", Author: "Author 22"},
		}

		resp := BooksListResponse{
			Books:  books,
			Total:  100,
			Limit:  10,
			Offset: 20,
		}

		assert.Equal(t, books, resp.Books)
		assert.Equal(t, 100, resp.Total)
		assert.Equal(t, 10, resp.Limit)
		assert.Equal(t, 20, resp.Offset)
	})
}
