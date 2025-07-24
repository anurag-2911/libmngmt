# Library Management System

[![CI/CD Pipeline](https://github.com/anurag-2911/libmngmt/actions/workflows/ci.yml/badge.svg)](https://github.com/anurag-2911/libmngmt/actions/workflows/ci.yml)
[![Security Scan](https://github.com/anurag-2911/libmngmt/actions/workflows/security.yml/badge.svg)](https://github.com/anurag-2911/libmngmt/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/anurag-2911/libmngmt)](https://goreportcard.com/repo### Local Monitoring Setup:\*\*

# Add Prometheus metrics (optional)

go get github.com/prometheus/client_golang

# Add structured logging

go get github.com/sirupsen/logrus

# Add distributed tracing

go get go.opentelemetry.io/otel

## Development Tools & Scripts

### Security Scanning

**Local Security Scan:**

# Run comprehensive security scan

chmod +x scripts/security-scan.sh
./scripts/security-scan.sh

The security scanner checks for:

- **API keys and tokens**: AWS, GitHub, JWT, etc.
- **Database credentials**: Passwords, connection strings
- **Private keys**: SSH, SSL, PGP keys
- **URLs with credentials**: HTTP basic auth, etc.

**Integration Testing:**

# API integration tests

chmod +x test_api.sh
./test_api.sh

# Rate limiting tests

chmod +x test_rate_limit.sh  
./test_rate_limit.sh

### Development Workflow

1. **Setup Development Environment:**

   git clone https://github.com/anurag-2911/libmngmt.git
   cd libmngmt
   cp .env.example .env
   docker-compose up -d

2. **Run Tests:**

   # Unit tests

   go test ./...

   # Integration tests

   ./test_api.sh

   # Security scan

   ./scripts/security-scan.sh

3. **Code Quality:**

   # Linting

   golangci-lint run

   # Format code

   go fmt ./...

   # Dependency updates

   go mod tidy

### Project Structure Details

libmngmt/
├── 📁 .github/
│ ├── 🔄 workflows/ # CI/CD automation
│ ├── 📋 ISSUE*TEMPLATE/ # Issue templates
│ └── 📋 PULL_REQUEST_TEMPLATE.md
├── 📁 cmd/api/ # Application entry point
├── 📁 internal/ # Private application code
│ ├── config/ # Configuration management
│ ├── database/ # Database layer
│ ├── handlers/ # HTTP handlers
│ ├── middleware/ # HTTP middleware
│ ├── models/ # Data models
│ ├── repository/ # Data access layer
│ └── service/ # Business logic
├── 📁 scripts/ # Automation scripts
│ └── security-scan.sh # Security vulnerability scanner
├── docker-compose\*.yml # Multi-environment Docker
├── Dockerfile # Secure container definition
├── init.sql # Database initialization
├── test*\*.sh # Integration test scripts
├── .golangci.yml # Linting configuration
├── Makefile # Build automation
└── API.md # API documentation
om/anurag-2911/libmngmt)
[![codecov](https://codecov.io/gh/anurag-2911/libmngmt/branch/main/graph/badge.svg)](https://codecov.io/gh/anurag-2911/libmngmt)
[![Go Reference](https://pkg.go.dev/badge/github.com/anurag-2911/libmngmt.svg)](https://pkg.go.dev/github.com/anurag-2911/libmngmt)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A production-grade REST API for library management built with Go, PostgreSQL, and Docker.

## Features

- Complete CRUD operations for books
- Clean Architecture (Repository, Service, Handler layers)
- Comprehensive book listing with filtering capabilities
- Dockerized with Docker Compose (development, staging, production)
- PostgreSQL database with proper schema and sample data
- RESTful API design with proper HTTP status codes
- Input validation and error handling
- Structured logging
- Environment-based configuration
- **CI/CD Pipeline** with GitHub Actions
- **Security Scanning** with automated vulnerability detection
- **Comprehensive Testing** (unit, integration, API testing)
- **Multi-stage Docker builds** with security best practices
- **Multi-environment deployments** (development, staging, production)
- **Security hardening** with secrets scanning and dependency updates
- **Rate limiting** and performance optimizations
- **Load balancing ready** with horizontal scaling support

## API Endpoints

| Method | Endpoint          | Description         |
| ------ | ----------------- | ------------------- |
| GET    | `/api/books`      | List all books      |
| GET    | `/api/books/{id}` | Get a specific book |
| POST   | `/api/books`      | Create a new book   |
| PUT    | `/api/books/{id}` | Update a book       |
| DELETE | `/api/books/{id}` | Delete a book       |

## Quick Start

### Docker Compose (Recommended)

1. **Clone the repository:**

   git clone https://github.com/anurag-2911/libmngmt.git
   cd libmngmt

2. **Start with Docker Compose:**

   # Development environment

   docker-compose up --build

   # Or for specific environments

   docker-compose -f docker-compose.staging.yml up --build
   docker-compose -f docker-compose.production.yml up --build

3. **The API will be available at:**
   - Development: `http://localhost:8080`
   - Staging: `http://localhost:8081`
   - Production: `http://localhost:8082`

### Environment Variables

Copy the appropriate environment file:

# For development

cp .env.example .env

# For staging

cp .env.staging.example .env.staging

# For production

cp .env.production.example .env.production

### Database Initialization

The database comes pre-loaded with sample data:

- **15 sample books** covering programming, DevOps, databases, and system design
- **Proper schema** with UUID primary keys, indexes, and triggers
- **Automatic migrations** on container startup

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL
- Docker & Docker Compose

### Local Development

1. Install dependencies:

   go mod download

2. Set up environment variables (copy `.env.example` to `.env`)

3. Run the application:

   go run cmd/api/main.go

## Architecture

The application follows Clean Architecture principles with additional DevOps and security layers:

├── cmd/api/ # Application entry point
├── internal/
│ ├── config/ # Configuration management  
│ ├── database/ # Database connection and migrations
│ ├── handlers/ # HTTP handlers (controllers)
│ ├── models/ # Domain models
│ ├── repository/ # Data access layer
│ ├── service/ # Business logic layer
│ └── middleware/ # HTTP middleware
├── .github/workflows/ # CI/CD pipelines
│ ├── ci.yml # Main CI/CD pipeline
│ └── security.yml # Security scanning workflow
├── scripts/ # Automation scripts
│ └── security-scan.sh # Security vulnerability scanner
├── docker-compose*.yml # Multi-environment Docker configurations
├── Dockerfile # Multi-stage secure container build
├── init.sql # Database initialization with sample data
└── test\_*.sh # Integration testing scripts

### CI/CD Pipeline Features

**Automated Testing & Quality:**

- Unit tests with coverage reporting
- Integration tests with real database
- API endpoint testing
- Go linting with 35+ rules (.golangci.yml)
- Security scanning (Gosec, Nancy, Trivy)
- Dependency vulnerability checks

**Multi-Environment Deployments:**

- **Staging**: Auto-deploy on `develop` branch
- **Production**: Manual approval for `main` branch
- **Docker**: Multi-stage builds with distroless images
- **Registry**: GitHub Container Registry integration

**Security & Compliance:**

- **Secret Scanning**: Automated detection of exposed credentials
- **Dependency Updates**: Dependabot security patches
- **Secure Builds**: Non-root containers, minimal attack surface
- **SECURITY.md**: Comprehensive security guidelines

## Testing

This project includes comprehensive testing at multiple levels to ensure reliability and correctness.

### Unit Tests

The application has 100% test coverage across all layers:

#### Run All Tests

# Run all tests

make test

# Run tests with coverage

make test-coverage

# Run specific package tests

go test ./internal/handlers/...
go test ./internal/service/...
go test ./internal/repository/...

#### Test Coverage by Package

- **Handlers**: HTTP request/response testing with mocks
- **Services**: Business logic validation and error handling
- **Repository**: Database operations with SQL mocks
- **Models**: Data validation and transformation

### Integration Tests

#### Docker Integration Testing

1. **Start the services:**

   docker-compose up -d

2. **Run the comprehensive API test suite:**

   chmod +x test_api.sh
   ./test_api.sh

#### Manual API Testing with curl

Test all endpoints manually:

**1. Health Check:**

curl -i http://localhost:8080/api/health

**2. Create a Book:**

curl -i -X POST http://localhost:8080/api/books \
 -H "Content-Type: application/json" \
 -d '{
"title": "The Hobbit",
"author": "J.R.R. Tolkien",
"isbn": "9780547928210",
"publisher": "Houghton Mifflin",
"genre": "Fantasy",
"published_at": "1937-09-21T00:00:00Z",
"pages": 366,
"language": "English"
}'

**3. Get All Books:**

curl -i http://localhost:8080/api/books

**4. Get Book by ID:**

curl -i http://localhost:8080/api/books/{book-id}

**5. Filter Books:**

# Filter by author (case-insensitive)

curl -i "http://localhost:8080/api/books?author=tolkien"

# Filter by genre

curl -i "http://localhost:8080/api/books?genre=fantasy"

# Filter by availability

curl -i "http://localhost:8080/api/books?available=true"

# Pagination

curl -i "http://localhost:8080/api/books?limit=10&offset=0"

**6. Update a Book:**

curl -i -X PUT http://localhost:8080/api/books/{book-id} \
 -H "Content-Type: application/json" \
 -d '{"available": false}'

**7. Delete a Book:**

curl -i -X DELETE http://localhost:8080/api/books/{book-id}

### Error Testing

The API properly handles various error scenarios:

**Invalid UUID:**

curl -i http://localhost:8080/api/books/invalid-uuid

# Returns: 400 Bad Request

# {"error":"Invalid book ID","message":"ID must be a valid UUID"}

**Book Not Found:**

curl -i http://localhost:8080/api/books/00000000-0000-0000-0000-000000000000

# Returns: 404 Not Found

# {"error":"Book not found","message":"failed to get book: book not found"}

**Duplicate ISBN:**

curl -i -X POST http://localhost:8080/api/books \
 -H "Content-Type: application/json" \
 -d '{"title":"Duplicate","author":"Author","isbn":"9780547928210"}'

# Returns: 409 Conflict

# {"error":"Duplicate resource","message":"book with ISBN 9780547928210 already exists"}

**Validation Errors:**

curl -i -X POST http://localhost:8080/api/books \
 -H "Content-Type: application/json" \
 -d '{"author":"Author","isbn":"9781234567890"}'

# Returns: 400 Bad Request

# {"error":"Validation error","message":"title is required"}

**Invalid JSON:**

curl -i -X POST http://localhost:8080/api/books \
 -H "Content-Type: application/json" \
 -d '{ invalid json'

# Returns: 400 Bad Request

# {"error":"Invalid JSON","message":"invalid character 'i' looking for beginning of object key string"}

### Test Files

- `test_api.sh` - Comprehensive API integration test suite
- `test_rate_limit.sh` - Rate limiting and performance testing
- `internal/handlers/*_test.go` - HTTP handler unit tests
- `internal/service/*_test.go` - Business logic unit tests
- `internal/repository/*_test.go` - Database layer unit tests
- `integration_test.go` - End-to-end integration tests

### Expected Response Formats

**Success Response:**
json
{
"message": "Book created successfully",
"data": {
"id": "uuid",
"title": "Book Title",
"author": "Author Name",
...
}
}

**Error Response:**
json
{
"error": "Error Type",
"message": "Detailed error message",
"code": 400
}

**List Response:**
json
{
"message": "Books retrieved successfully",
"data": {
"books": [...],
"total": 10,
"limit": 50,
"offset": 0
}
}

### Test Results Summary

When all tests pass, you should see:

# Unit Tests

$ make test

# Integration Tests

$ ./test_api.sh

### Performance & Load Testing

The API includes built-in rate limiting (100 concurrent requests) and can be tested with:

# Test rate limiting

for i in {1..110}; do curl -s http://localhost:8080/api/books > /dev/null & done

# Load testing with wrk (if installed)

wrk -t12 -c400 -d30s http://localhost:8080/api/books

# Database performance testing

docker-compose exec postgres pgbench -c 10 -T 60 -h localhost -U libuser libmngmt

## CI/CD & DevOps

### GitHub Actions Workflows

**Main CI/CD Pipeline (`.github/workflows/ci.yml`):**

1. **Test Stage**:

   - Unit tests across all packages
   - Integration tests with PostgreSQL
   - Go linting with golangci-lint
   - Code coverage reporting

2. **Security Stage**:

   - GoSec security scanning
   - Nancy dependency vulnerability checks
   - Trivy container image scanning
   - Custom secret detection script

3. **Build Stage**:

   - Multi-stage Docker image builds
   - Image optimization and security hardening
   - Push to GitHub Container Registry
   - Build artifacts with version tagging

4. **Deploy Staging**:

   - Auto-deploy on `develop` branch pushes
   - Environment variable validation
   - Health check verification
   - Rollback on failure

5. **Deploy Production**:
   - Manual approval required
   - Triggered on `main` branch
   - Blue-green deployment strategy
   - Production smoke tests

**Security Scanning Workflow (`.github/workflows/security.yml`):**

- **Scheduled scans**: Weekly security audits
- **Dependency scanning**: Automated vulnerability detection
- **SARIF reporting**: Integration with GitHub Security tab
- **Alert notifications**: Slack/email notifications for critical findings

### Security Features

**Repository Security:**

- **Secret scanning**: Prevents accidental credential commits
- **.gitignore hardening**: Comprehensive exclusion patterns
- **Environment templates**: Safe configuration examples
- **Security documentation**: SECURITY.md with best practices

**Application Security:**

- **Input validation**: Comprehensive request validation
- **SQL injection prevention**: Parameterized queries
- **Rate limiting**: 100 concurrent request limit
- **Security headers**: HTTP security headers implementation

**Container Security:**

- **Distroless images**: Minimal attack surface
- **Non-root execution**: Unprivileged container user
- **Vulnerability scanning**: Automated image security checks
- **Multi-stage builds**: Build-time security separation

### Deployment Environments

| Environment     | Trigger           | URL                    | Database         | Features                  |
| --------------- | ----------------- | ---------------------- | ---------------- | ------------------------- |
| **Development** | Local/Manual      | `localhost:8080`       | Local PostgreSQL | Hot reload, debug logging |
| **Staging**     | Push to `develop` | `staging.libmngmt.com` | Staging DB       | Production-like, testing  |
| **Production**  | Manual approval   | `api.libmngmt.com`     | Production DB    | Optimized, monitoring     |

### Monitoring & Observability

**Available Integrations:**
yaml

# Add to your workflow for enhanced monitoring

- name: Performance Monitoring
  uses: actions/performance-monitor@v1
- name: Error Tracking  
  uses: sentry/action-release@v1
- name: Uptime Monitoring
  uses: upptime/uptime-monitor@v1

**Local Monitoring Setup:**

# Add Prometheus metrics (optional)

go get github.com/prometheus/client_golang

# Add structured logging

go get github.com/sirupsen/logrus

# Add distributed tracing

go get go.opentelemetry.io/otel

## Performance Analysis & Scaling

### Current Performance Characteristics

**Single Instance Capacity:**

- ~1,000-5,000 requests/second (depending on query complexity)
- Memory usage: ~50-100MB at rest
- Database connections: Pooled (default 25 max connections)
- Rate limiting: 100 concurrent requests per instance

**For Millions of Requests Per Day:**

- 1M requests/day = ~12 requests/second (manageable)
- 10M requests/day = ~115 requests/second (requires optimization)
- 100M requests/day = ~1,157 requests/second (requires scaling)

### Recommended Performance Improvements

#### 1. **Redis Caching Layer**

_Essential for high-traffic scenarios_

Added Redis for:

- **Query Result Caching**: Cache frequent book searches
- **Session Management**: If authentication is added
- **Rate Limiting**: Distributed rate limiting across instances

yaml

# docker-compose.yml addition

redis:
image: redis:7-alpine
container_name: libmngmt_redis
ports: - "6379:6379"
volumes: - redis_data:/data
networks: - libmngmt_network

### Monitoring & Observability

**Add these for production:**

# Prometheus metrics

go get github.com/prometheus/client_golang

# Structured logging

go get github.com/sirupsen/logrus

# Distributed tracing

go get go.opentelemetry.io/otel

**Key Metrics to Monitor:**

- Request rate (req/sec)
- Response time (p95, p99)
- Database connection pool usage
- Redis hit/miss ratio
- Memory usage
- CPU utilization

### Performance Testing

# Load testing with Apache Bench

ab -n 10000 -c 100 http://localhost:8080/api/books

# Advanced load testing with hey

hey -n 50000 -c 200 -m GET http://localhost:8080/api/books

# Database stress testing

pgbench -c 10 -T 60 -h localhost -U libuser libmngmt

### Cost-Effective Scaling Path

1. **Phase 1** (1M-10M requests/day): Add Redis ($10-20/month)
2. **Phase 2** (10M-50M requests/day): Load balancer + 3 instances ($100-200/month)
3. **Phase 3** (50M+ requests/day): Auto-scaling + managed services ($500+/month)

**Conclusion:** The current architecture can handle 1M requests/day with Redis caching. For 10M+ requests/day, horizontal scaling with load balancing becomes essential.



### Development Process

1. **Fork & Clone:**

   git fork https://github.com/anurag-2911/libmngmt.git
   git clone https://github.com/YOUR_USERNAME/libmngmt.git
   cd libmngmt

2. **Setup Development Environment:**

   cp .env.example .env
   docker-compose up -d
   go mod download

3. **Create Feature Branch:**

   git checkout -b feature/your-feature-name

4. **Make Changes & Test:**

   # Run tests

   go test ./...
   ./test_api.sh

   # Security scan

   ./scripts/security-scan.sh

   # Linting

   golangci-lint run

5. **Submit Pull Request:**
   - Ensure all tests pass
   - Update documentation if needed
   - Follow conventional commit messages
   - Add appropriate labels

### Code Standards

- **Go Style**: Follow effective Go patterns
- **Testing**: Maintain 100% test coverage
- **Security**: Never commit secrets or credentials
- **Documentation**: Update README and API docs
- **CI/CD**: Ensure all pipeline checks pass

## License

### Docker Test Environment

The Docker setup provides a complete isolated test environment with multiple configurations:

#### Available Docker Compositions

**Development (`docker-compose.yml`):**

- PostgreSQL 15 with sample data
- Redis for caching (when enabled)
- API server with hot-reload support
- Development-friendly logging

**Staging (`docker-compose.staging.yml`):**

- Production-like environment
- Environment variable validation
- Performance monitoring
- Staging-specific configurations

**Production (`docker-compose.production.yml`):**

- Optimized for production workloads
- Security hardening enabled
- Resource limits and health checks
- Production logging and monitoring

#### Container Features

**Multi-stage Dockerfile:**

- **Build stage**: Go compilation with dependency caching
- **Production stage**: Distroless base image for security
- **Non-root user**: Runs as unprivileged user (UID 10001)
- **Minimal size**: <20MB final image
- **Security scanning**: Integrated vulnerability detection

**Database Container:**

- **PostgreSQL 15**: Latest stable version with Alpine base
- **Sample data**: 15 pre-loaded books for testing
- **Health checks**: Automatic service health monitoring
- **Persistent storage**: Data volumes for development persistence
- **Auto-initialization**: Runs init.sql on first startup

#### Quick Commands

# Start development environment

docker-compose up -d

# Check service health

docker-compose ps

# View service logs

docker-compose logs api
docker-compose logs postgres

# Run tests in containers

docker-compose exec api go test ./...

# Database operations

docker-compose exec postgres psql -U libuser -d libmngmt

# Clean up

docker-compose down --volumes

#### Environment Variables

Each environment uses different variable files:

# Development

.env # Local development overrides
.env.example # Template with safe defaults

# Staging

.env.staging.example # Staging-specific configuration

# Production

.env.production.example # Production-ready settings

## License
