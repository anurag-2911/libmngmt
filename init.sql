-- Initialize database with schema and sample data
-- This file will be executed when the PostgreSQL container starts

-- Enable UUID extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create books table (matching application schema)
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

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_genre ON books(genre);
CREATE INDEX IF NOT EXISTS idx_books_available ON books(available);
CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn);

-- Create a trigger to automatically update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_books_updated_at BEFORE UPDATE ON books
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample books (updated with correct schema)
INSERT INTO books (title, author, isbn, publisher, genre, published_at, pages, language, available) VALUES
('The Go Programming Language', 'Alan Donovan, Brian Kernighan', '978-0134190440', 'Addison-Wesley', 'Programming', '2015-10-26'::timestamp, 380, 'English', true),
('Clean Architecture', 'Robert C. Martin', '978-0134494166', 'Prentice Hall', 'Software Engineering', '2017-09-10'::timestamp, 432, 'English', true),
('Docker Deep Dive', 'Nigel Poulton', '978-1521822807', 'Independently Published', 'DevOps', '2017-07-14'::timestamp, 280, 'English', true),
('PostgreSQL: Up and Running', 'Regina Obe, Leo Hsu', '978-1449373184', 'O''Reilly Media', 'Database', '2014-10-15'::timestamp, 270, 'English', true),
('System Design Interview', 'Alex Xu', '978-1736049129', 'Independently Published', 'System Design', '2020-06-30'::timestamp, 322, 'English', false),
('Microservices Patterns', 'Chris Richardson', '978-1617294549', 'Manning Publications', 'Architecture', '2018-10-19'::timestamp, 520, 'English', true),
('Building Microservices', 'Sam Newman', '978-1491950357', 'O''Reilly Media', 'Architecture', '2015-02-20'::timestamp, 280, 'English', true),
('Kubernetes in Action', 'Marko Luksa', '978-1617293726', 'Manning Publications', 'DevOps', '2017-12-17'::timestamp, 624, 'English', true),
('Effective Go', 'Dave Cheney', '978-1234567890', 'Tech Books', 'Programming', '2020-01-15'::timestamp, 200, 'English', true),
('Database Internals', 'Alex Petrov', '978-1492040347', 'O''Reilly Media', 'Database', '2019-09-27'::timestamp, 373, 'English', true),
('Site Reliability Engineering', 'Various Authors', '978-1491929124', 'O''Reilly Media', 'Operations', '2016-03-23'::timestamp, 552, 'English', true),
('Designing Data-Intensive Applications', 'Martin Kleppmann', '978-1449373320', 'O''Reilly Media', 'System Design', '2017-03-16'::timestamp, 616, 'English', true),
('The DevOps Handbook', 'Gene Kim', '978-1942788003', 'IT Revolution Press', 'DevOps', '2016-10-06'::timestamp, 480, 'English', false),
('Prometheus: Up & Running', 'Brian Brazil', '978-1492034148', 'O''Reilly Media', 'Monitoring', '2018-07-17'::timestamp, 396, 'English', true),
('Container Security', 'Liz Rice', '978-1492056706', 'O''Reilly Media', 'Security', '2020-04-06'::timestamp, 234, 'English', true);

-- Create a view for available books
CREATE OR REPLACE VIEW available_books AS
   SELECT 
       id,
       title,
       author,
       isbn,
       publisher,
       genre,
       published_at,
       pages,
       language,
       created_at,
       updated_at
   FROM books 
   WHERE available = true;-- Create a statistics view for administrative purposes
CREATE OR REPLACE VIEW book_statistics AS
   SELECT 
       COUNT(*) as total_books,
       COUNT(*) FILTER (WHERE available = true) as available_books,
       COUNT(*) FILTER (WHERE available = false) as borrowed_books,
       COUNT(DISTINCT author) as unique_authors,
       COUNT(DISTINCT genre) as unique_genres,
       AVG(pages) as average_pages,
       MIN(EXTRACT(YEAR FROM published_at)) as oldest_book_year,
       MAX(EXTRACT(YEAR FROM published_at)) as newest_book_year
   FROM books;

-- Grant necessary permissions (if using specific user)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO libuser;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO libuser;
-- GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO libuser;

-- Display initial statistics
SELECT 'Database initialized successfully!' as status;
SELECT * FROM book_statistics;
