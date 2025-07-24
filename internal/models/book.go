package models

import (
	"time"

	"github.com/google/uuid"
)

// Book represents a book in the library
type Book struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Author      string    `json:"author" db:"author"`
	ISBN        string    `json:"isbn" db:"isbn"`
	Publisher   string    `json:"publisher" db:"publisher"`
	Genre       string    `json:"genre" db:"genre"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	Pages       int       `json:"pages" db:"pages"`
	Language    string    `json:"language" db:"language"`
	Available   bool      `json:"available" db:"available"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateBookRequest represents the request body for creating a book
type CreateBookRequest struct {
	Title       string    `json:"title" validate:"required,min=1,max=255"`
	Author      string    `json:"author" validate:"required,min=1,max=255"`
	ISBN        string    `json:"isbn" validate:"required,min=10,max=17"`
	Publisher   string    `json:"publisher" validate:"max=255"`
	Genre       string    `json:"genre" validate:"max=100"`
	PublishedAt time.Time `json:"published_at"`
	Pages       int       `json:"pages" validate:"min=1"`
	Language    string    `json:"language" validate:"max=50"`
}

// UpdateBookRequest represents the request body for updating a book
type UpdateBookRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Author      *string    `json:"author,omitempty" validate:"omitempty,min=1,max=255"`
	ISBN        *string    `json:"isbn,omitempty" validate:"omitempty,min=10,max=17"`
	Publisher   *string    `json:"publisher,omitempty" validate:"omitempty,max=255"`
	Genre       *string    `json:"genre,omitempty" validate:"omitempty,max=100"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Pages       *int       `json:"pages,omitempty" validate:"omitempty,min=1"`
	Language    *string    `json:"language,omitempty" validate:"omitempty,max=50"`
	Available   *bool      `json:"available,omitempty"`
}

// BookFilter represents filters for listing books
type BookFilter struct {
	Author    string `json:"author,omitempty"`
	Genre     string `json:"genre,omitempty"`
	Language  string `json:"language,omitempty"`
	Available *bool  `json:"available,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// BooksListResponse represents the response for listing books
type BooksListResponse struct {
	Books  []Book `json:"books"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
