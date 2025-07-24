package repository

import (
	"database/sql"
	"fmt"
	"libmngmt/internal/database"
	"libmngmt/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBookRepository(t *testing.T) {
	t.Run("create new book repository", func(t *testing.T) {
		db, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)
		assert.NotNil(t, repo)
		assert.IsType(t, &bookRepository{}, repo)
	})
}

func TestBookRepository_Create(t *testing.T) {
	t.Run("create book successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id := uuid.New()
		now := time.Now()
		req := &models.CreateBookRequest{
			Title:       "Test Book",
			Author:      "Test Author",
			ISBN:        "9781234567890",
			Publisher:   "Test Publisher",
			Genre:       "Fiction",
			PublishedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Pages:       300,
			Language:    "English",
		}

		expectedQuery := `INSERT INTO books`
		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(
				sqlmock.AnyArg(), // id (UUID)
				req.Title,
				req.Author,
				req.ISBN,
				req.Publisher,
				req.Genre,
				req.PublishedAt,
				req.Pages,
				req.Language,
				true,             // available (default)
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at",
			}).AddRow(
				id, now, now,
			))

		book, err := repo.Create(req)

		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, req.Title, book.Title)
		assert.Equal(t, req.Author, book.Author)
		assert.Equal(t, req.ISBN, book.ISBN)
		assert.Equal(t, req.Publisher, book.Publisher)
		assert.Equal(t, req.Genre, book.Genre)
		assert.Equal(t, req.Pages, book.Pages)
		assert.Equal(t, req.Language, book.Language)
		assert.True(t, book.Available)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create book with database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		req := &models.CreateBookRequest{
			Title:    "Test Book",
			Author:   "Test Author",
			ISBN:     "9781234567890",
			Pages:    300,
			Language: "English",
		}

		expectedQuery := `INSERT INTO books`
		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(sqlmock.AnyArg(), req.Title, req.Author, req.ISBN, req.Publisher,
				req.Genre, req.PublishedAt, req.Pages, req.Language, true,
				sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("database error"))

		book, err := repo.Create(req)

		assert.Error(t, err)
		assert.Nil(t, book)
		assert.Contains(t, err.Error(), "database error")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBookRepository_GetByID(t *testing.T) {
	t.Run("get book by ID successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id := uuid.New()
		now := time.Now()
		publishedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		expectedQuery := `SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at FROM books WHERE id = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "author", "isbn", "publisher", "genre",
				"published_at", "pages", "language", "available", "created_at", "updated_at",
			}).AddRow(
				id, "Test Book", "Test Author", "9781234567890", "Test Publisher", "Fiction",
				publishedAt, 300, "English", true, now, now,
			))

		book, err := repo.GetByID(id)

		assert.NoError(t, err)
		assert.NotNil(t, book)
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

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get book by ID not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id := uuid.New()

		expectedQuery := `SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at FROM books WHERE id = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		book, err := repo.GetByID(id)

		assert.Error(t, err)
		assert.Nil(t, book)
		assert.Contains(t, err.Error(), "book not found")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBookRepository_Update(t *testing.T) {
	t.Run("update book successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id := uuid.New()
		now := time.Now()
		newTitle := "Updated Title"
		newAuthor := "Updated Author"

		req := &models.UpdateBookRequest{
			Title:  &newTitle,
			Author: &newAuthor,
		}

		// First expect GetByID call
		selectQuery := `SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at FROM books WHERE id = \$1`
		mock.ExpectQuery(selectQuery).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "author", "isbn", "publisher", "genre",
				"published_at", "pages", "language", "available", "created_at", "updated_at",
			}).AddRow(
				id, "Original Title", "Original Author", "9781234567890", "Test Publisher", "Fiction",
				time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), 300, "English", true, now, now,
			))

		// Then expect UPDATE query
		expectedQuery := `UPDATE books SET title = \$1, author = \$2, updated_at = \$3 WHERE id = \$4`
		mock.ExpectQuery(expectedQuery).
			WithArgs(newTitle, newAuthor, sqlmock.AnyArg(), id).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "author", "isbn", "publisher", "genre",
				"published_at", "pages", "language", "available", "created_at", "updated_at",
			}).AddRow(
				id, newTitle, newAuthor, "9781234567890", "Test Publisher", "Fiction",
				time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), 300, "English", true, now, now,
			))

		book, err := repo.Update(id, req)

		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, id, book.ID)
		assert.Equal(t, newTitle, book.Title)
		assert.Equal(t, newAuthor, book.Author)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update book with empty request", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id := uuid.New()
		now := time.Now()
		req := &models.UpdateBookRequest{}

		// Expect GetByID call since Update calls GetByID first
		selectQuery := `SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at FROM books WHERE id = \$1`
		mock.ExpectQuery(selectQuery).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "author", "isbn", "publisher", "genre",
				"published_at", "pages", "language", "available", "created_at", "updated_at",
			}).AddRow(
				id, "Original Title", "Original Author", "9781234567890", "Test Publisher", "Fiction",
				time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), 300, "English", true, now, now,
			))

		book, err := repo.Update(id, req)

		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, id, book.ID)
		assert.Equal(t, "Original Title", book.Title) // Should return unchanged book

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBookRepository_Delete(t *testing.T) {
	t.Run("delete book successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id := uuid.New()

		expectedQuery := `DELETE FROM books WHERE id = \$1`
		mock.ExpectExec(expectedQuery).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Delete(id)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete book not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id := uuid.New()

		expectedQuery := `DELETE FROM books WHERE id = \$1`
		mock.ExpectExec(expectedQuery).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = repo.Delete(id)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "book not found")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBookRepository_GetAll(t *testing.T) {
	t.Run("get all books successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		id1 := uuid.New()
		id2 := uuid.New()
		now := time.Now()

		filter := models.BookFilter{
			Limit:  10,
			Offset: 0,
		}

		// Count query
		expectedCountQuery := `SELECT COUNT\(\*\) FROM books`
		mock.ExpectQuery(expectedCountQuery).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Main query
		expectedQuery := `SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at FROM books ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`
		mock.ExpectQuery(expectedQuery).
			WithArgs(10, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "author", "isbn", "publisher", "genre",
				"published_at", "pages", "language", "available", "created_at", "updated_at",
			}).
				AddRow(id1, "Book 1", "Author 1", "9781111111111", "Publisher 1", "Fiction",
					time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), 200, "English", true, now, now).
				AddRow(id2, "Book 2", "Author 2", "9782222222222", "Publisher 2", "Non-Fiction",
					time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), 250, "English", false, now, now))

		books, total, err := repo.GetAll(filter)

		assert.NoError(t, err)
		assert.NotNil(t, books)
		assert.Len(t, books, 2)
		assert.Equal(t, 2, total)

		// Check first book
		assert.Equal(t, id1, books[0].ID)
		assert.Equal(t, "Book 1", books[0].Title)
		assert.Equal(t, "Author 1", books[0].Author)

		// Check second book
		assert.Equal(t, id2, books[1].ID)
		assert.Equal(t, "Book 2", books[1].Title)
		assert.Equal(t, "Author 2", books[1].Author)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get all books with filters", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		available := true
		filter := models.BookFilter{
			Author:    "tolkien",
			Genre:     "fantasy",
			Language:  "English",
			Available: &available,
			Limit:     5,
			Offset:    10,
		}

		// Count query with WHERE clause
		expectedCountQuery := `SELECT COUNT\(\*\) FROM books WHERE LOWER\(author\) LIKE LOWER\(\$1\) AND LOWER\(genre\) LIKE LOWER\(\$2\) AND LOWER\(language\) = LOWER\(\$3\) AND available = \$4`
		mock.ExpectQuery(expectedCountQuery).
			WithArgs("%tolkien%", "%fantasy%", "English", true).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Main query with WHERE clause
		expectedQuery := `SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at FROM books WHERE LOWER\(author\) LIKE LOWER\(\$1\) AND LOWER\(genre\) LIKE LOWER\(\$2\) AND LOWER\(language\) = LOWER\(\$3\) AND available = \$4 ORDER BY created_at DESC LIMIT \$5 OFFSET \$6`
		mock.ExpectQuery(expectedQuery).
			WithArgs("%tolkien%", "%fantasy%", "English", true, 5, 10).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "author", "isbn", "publisher", "genre",
				"published_at", "pages", "language", "available", "created_at", "updated_at",
			}).
				AddRow(uuid.New(), "The Hobbit", "J.R.R. Tolkien", "9780547928227", "Houghton Mifflin", "Fantasy",
					time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC), 310, "English", true, time.Now(), time.Now()))

		books, total, err := repo.GetAll(filter)

		assert.NoError(t, err)
		assert.NotNil(t, books)
		assert.Len(t, books, 1)
		assert.Equal(t, 1, total)

		assert.Equal(t, "The Hobbit", books[0].Title)
		assert.Equal(t, "J.R.R. Tolkien", books[0].Author)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get all books empty result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		filter := models.BookFilter{
			Limit:  10,
			Offset: 0,
		}

		// Count query
		expectedCountQuery := `SELECT COUNT\(\*\) FROM books`
		mock.ExpectQuery(expectedCountQuery).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Main query
		expectedQuery := `SELECT id, title, author, isbn, publisher, genre, published_at, pages, language, available, created_at, updated_at FROM books ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`
		mock.ExpectQuery(expectedQuery).
			WithArgs(10, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "author", "isbn", "publisher", "genre",
				"published_at", "pages", "language", "available", "created_at", "updated_at",
			}))

		books, total, err := repo.GetAll(filter)

		assert.NoError(t, err)
		assert.NotNil(t, books)
		assert.Len(t, books, 0)
		assert.Equal(t, 0, total)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBookRepository_ExistsByISBN(t *testing.T) {
	t.Run("ISBN exists", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		isbn := "9781234567890"

		expectedQuery := `SELECT COUNT\(\*\) FROM books WHERE isbn = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(isbn).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		exists, err := repo.ExistsByISBN(isbn, nil)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ISBN does not exist", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		isbn := "9781234567890"

		expectedQuery := `SELECT COUNT\(\*\) FROM books WHERE isbn = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(isbn).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := repo.ExistsByISBN(isbn, nil)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ISBN exists but excluded", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		dbWrapper := &database.DB{DB: db}
		repo := NewBookRepository(dbWrapper)

		isbn := "9781234567890"
		excludeID := uuid.New()

		expectedQuery := `SELECT COUNT\(\*\) FROM books WHERE isbn = \$1 AND id != \$2`
		mock.ExpectQuery(expectedQuery).
			WithArgs(isbn, excludeID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := repo.ExistsByISBN(isbn, &excludeID)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
