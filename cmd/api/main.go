package main

import (
	"context"
	"fmt"
	"libmngmt/internal/cache"
	"libmngmt/internal/config"
	"libmngmt/internal/database"
	"libmngmt/internal/handlers"
	"libmngmt/internal/middleware"
	"libmngmt/internal/repository"
	"libmngmt/internal/service"
	"libmngmt/internal/workers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration with proper error handling
	cfg, err := config.LoadWithValidation()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	bookRepo := repository.NewBookRepository(db)

	// Initialize Redis cache
	var bookCache *cache.BookCache
	if cfg.Redis.Enabled {
		log.Println("Initializing Redis cache...")
		addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
		redisCache := cache.NewRedisCache(addr, cfg.Redis.Password, cfg.Redis.DB)

		// Test Redis connection
		ctx := context.Background()
		if err := redisCache.Ping(ctx); err != nil {
			log.Printf("Failed to connect to Redis: %v, falling back to in-memory cache", err)
			bookCache = cache.NewBookCache(5*time.Minute, time.Minute, nil)
		} else {
			bookCache = cache.NewBookCache(5*time.Minute, time.Minute, redisCache)
			log.Println("Redis cache initialized successfully")
		}
	} else {
		log.Println("Redis disabled, using in-memory cache only")
		bookCache = cache.NewBookCache(5*time.Minute, time.Minute, nil)
	}

	// Create enhanced components
	workerPool := workers.NewBookProcessor(10, 100)

	// Initialize enhanced services
	bookService := service.NewBookService(bookRepo, bookCache, workerPool)

	// Initialize enhanced handlers
	bookHandler := handlers.NewBookHandler(bookService)

	// Setup routes
	router := setupRoutes(bookHandler)

	// Setup middleware
	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.JSONMiddleware)

	// Setup HTTP server with graceful shutdown
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", addr)
		log.Printf("Goroutines at startup: %d", runtime.NumGoroutine())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-c

	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown components gracefully
	if err := bookHandler.Shutdown(ctx); err != nil {
		log.Printf("Handler shutdown error: %v", err)
	}

	if err := workerPool.Shutdown(ctx); err != nil {
		log.Printf("Worker pool shutdown error: %v", err)
	}

	// Shutdown cache
	if bookCache != nil {
		bookCache.Shutdown()
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func setupRoutes(bookHandler *handlers.BookHandler) *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Book routes
	api.HandleFunc("/books", bookHandler.GetBooks).Methods("GET")
	api.HandleFunc("/books", bookHandler.CreateBook).Methods("POST")
	api.HandleFunc("/books/{id}", bookHandler.GetBook).Methods("GET")
	api.HandleFunc("/books/{id}", bookHandler.UpdateBook).Methods("PUT")
	api.HandleFunc("/books/{id}", bookHandler.DeleteBook).Methods("DELETE")
	api.HandleFunc("/books/bulk", bookHandler.BulkCreateBooks).Methods("POST")
	api.HandleFunc("/books/metrics", bookHandler.GetMetrics).Methods("GET")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"healthy","goroutines":%d}`, runtime.NumGoroutine())))
	}).Methods("GET")

	// API documentation endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message": "Library Management API - Production Ready with Go Concurrency",
			"version": "2.0.0",
			"endpoints": {
				"books": {
					"GET /api/books": "Get all books with filtering and caching",
					"POST /api/books": "Create a book with concurrent validation",
					"GET /api/books/{id}": "Get a book by ID with caching",
					"PUT /api/books/{id}": "Update a book with validation",
					"DELETE /api/books/{id}": "Delete a book",
					"POST /api/books/bulk": "Bulk create books with worker pool",
					"GET /api/books/metrics": "Get performance metrics"
				},
				"utility": {
					"GET /health": "Health check with goroutine count"
				}
			},
			"features": [
				"Thread-safe caching with TTL",
				"Worker pool for bulk operations",
				"Rate limiting with channels",
				"Context-based timeouts",
				"Graceful shutdown handling",
				"Concurrent request processing",
				"Performance metrics tracking",
				"Enhanced error handling",
				"Input validation and sanitization"
			]
		}`))
	}).Methods("GET")

	return router
}
