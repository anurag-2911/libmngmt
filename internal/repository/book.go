package repository

import (
	"database/sql"
	"fmt"
	"libmngmt/internal/database"
	"libmngmt/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BookRepository defines the interface for book data operations
type BookRepository interface {
	Create(book *models.CreateBookRequest) (*models.Book, error)
	GetByID(id uuid.UUID) (*models.Book, error)
	GetAll(filter models.BookFilter) ([]models.Book, int, error)
	Update(id uuid.UUID, book *models.UpdateBookRequest) (*models.Book, error)
	Delete(id uuid.UUID) error
	ExistsByISBN(isbn string, excludeID *uuid.UUID) (bool, error)
}

// bookRepository implements BookRepository interface
type bookRepository struct {
	db *database.DB
}

// NewBookRepository creates a new book repository
func NewBookRepository(db *database.DB) BookRepository {
	return &bookRepository{db: db}
}

// Create creates a new book
func (r *bookRepository) Create(req *models.CreateBookRequest) (*models.Book, error) {
	book := &models.Book{
		ID:          uuid.New(),
		Title:       req.Title,
		Author:      req.Author,
		ISBN:        req.ISBN,
		Publisher:   req.Publisher,
		Genre:       req.Genre,
		PublishedAt: req.PublishedAt,
		Pages:       req.Pages,
		Language:    req.Language,
		Available:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if book.Language == "" {
		book.Language = "English"
	}

	query := `
		INSERT INTO books (id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		book.ID, book.Title, book.Author, book.ISBN, book.Publisher, book.Genre,
		book.PublishedAt, book.Pages, book.Language, book.Available, book.CreatedAt, book.UpdatedAt,
	).Scan(&book.ID, &book.CreatedAt, &book.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	return book, nil
}

// GetByID retrieves a book by its ID
func (r *bookRepository) GetByID(id uuid.UUID) (*models.Book, error) {
	book := &models.Book{}
	query := `
		SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at
		FROM books
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Publisher, &book.Genre,
		&book.PublishedAt, &book.Pages, &book.Language, &book.Available, &book.CreatedAt, &book.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("book not found")
		}
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	return book, nil
}

// GetAll retrieves all books with optional filtering
func (r *bookRepository) GetAll(filter models.BookFilter) ([]models.Book, int, error) {
	books := make([]models.Book, 0) // Initialize as empty slice, not nil slice
	var total int

	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argCount := 0

	if filter.Author != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(author) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+filter.Author+"%")
	}

	if filter.Genre != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(genre) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+filter.Genre+"%")
	}

	if filter.Language != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(language) = LOWER($%d)", argCount))
		args = append(args, filter.Language)
	}

	if filter.Available != nil {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("available = $%d", argCount))
		args = append(args, *filter.Available)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM books %s", whereClause)
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count books: %w", err)
	}

	// Set default pagination
	if filter.Limit <= 0 {
		filter.Limit = 50 // Default limit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Get books
	query := fmt.Sprintf(`
		SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at
		FROM books %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount+1, argCount+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query books: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var book models.Book
		err := rows.Scan(
			&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Publisher, &book.Genre,
			&book.PublishedAt, &book.Pages, &book.Language, &book.Available, &book.CreatedAt, &book.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan book: %w", err)
		}
		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return books, total, nil
}

// Update updates a book by its ID
func (r *bookRepository) Update(id uuid.UUID, req *models.UpdateBookRequest) (*models.Book, error) {
	// First, get the current book
	currentBook, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	var setParts []string
	var args []interface{}
	argCount := 0

	if req.Title != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("title = $%d", argCount))
		args = append(args, *req.Title)
	}

	if req.Author != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("author = $%d", argCount))
		args = append(args, *req.Author)
	}

	if req.ISBN != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("isbn = $%d", argCount))
		args = append(args, *req.ISBN)
	}

	if req.Publisher != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("publisher = $%d", argCount))
		args = append(args, *req.Publisher)
	}

	if req.Genre != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("genre = $%d", argCount))
		args = append(args, *req.Genre)
	}

	if req.PublishedAt != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("published_at = $%d", argCount))
		args = append(args, *req.PublishedAt)
	}

	if req.Pages != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("pages = $%d", argCount))
		args = append(args, *req.Pages)
	}

	if req.Language != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("language = $%d", argCount))
		args = append(args, *req.Language)
	}

	if req.Available != nil {
		argCount++
		setParts = append(setParts, fmt.Sprintf("available = $%d", argCount))
		args = append(args, *req.Available)
	}

	if len(setParts) == 0 {
		return currentBook, nil // No updates requested
	}

	// Add updated_at
	argCount++
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())

	// Add ID for WHERE clause
	argCount++
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE books 
		SET %s
		WHERE id = $%d
		RETURNING id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at
	`, strings.Join(setParts, ", "), argCount)

	book := &models.Book{}
	err = r.db.QueryRow(query, args...).Scan(
		&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Publisher, &book.Genre,
		&book.PublishedAt, &book.Pages, &book.Language, &book.Available, &book.CreatedAt, &book.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update book: %w", err)
	}

	return book, nil
}

// Delete deletes a book by its ID
func (r *bookRepository) Delete(id uuid.UUID) error {
	query := "DELETE FROM books WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("book not found")
	}

	return nil
}

// ExistsByISBN checks if a book with the given ISBN exists
func (r *bookRepository) ExistsByISBN(isbn string, excludeID *uuid.UUID) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM books WHERE isbn = $1"
	args := []interface{}{isbn}

	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, *excludeID)
	}

	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check ISBN existence: %w", err)
	}

	return count > 0, nil
}
