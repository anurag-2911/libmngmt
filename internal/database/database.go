package database

import (
	"database/sql"
	"fmt"
	"log"

	"libmngmt/internal/config"

	_ "github.com/lib/pq"
)

// DB wraps the sql.DB to provide additional functionality
type DB struct {
	*sql.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Migrate runs database migrations
func (db *DB) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS books (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		title VARCHAR(255) NOT NULL,
		author VARCHAR(255) NOT NULL,
		isbn VARCHAR(17) UNIQUE NOT NULL,
		publisher VARCHAR(255),
		genre VARCHAR(100),
		published_at TIMESTAMP,
		pages INTEGER,
		language VARCHAR(50) DEFAULT 'English',
		available BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
	CREATE INDEX IF NOT EXISTS idx_books_genre ON books(genre);
	CREATE INDEX IF NOT EXISTS idx_books_available ON books(available);
	CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn);

	-- Trigger to update updated_at timestamp
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	DROP TRIGGER IF EXISTS update_books_updated_at ON books;
	CREATE TRIGGER update_books_updated_at
		BEFORE UPDATE ON books
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed")
	return nil
}
