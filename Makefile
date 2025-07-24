.PHONY: build run test clean docker-build docker-run docker-down deps

# Default target
all: build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the application
build:
	go build -o bin/api cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-build:
	docker-compose build

docker-run:
	docker-compose up -d

docker-logs:
	docker-compose logs -f

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v
	docker system prune -f

# Development helpers
dev-db:
	docker run --name libmngmt-postgres -e POSTGRES_USER=libuser -e POSTGRES_PASSWORD=changeme_dev_only -e POSTGRES_DB=libmngmt -p 5432:5432 -d postgres:15-alpine

dev-db-stop:
	docker stop libmngmt-postgres
	docker rm libmngmt-postgres

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Security scan (requires gosec)
security:
	gosec ./...

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
