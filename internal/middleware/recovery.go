package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.Printf("Panic recovered: %v\nStack trace:\n%s", err, debug.Stack())

				// Return a clean error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResponse := map[string]interface{}{
					"error":   "Internal server error",
					"message": "An unexpected error occurred. Please try again later.",
				}

				if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
					log.Printf("Failed to encode error response: %v", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
