package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs request and response", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Wrap with logging middleware
		handler := LoggingMiddleware(testHandler)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")

		recorder := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(recorder, req)

		// Verify response
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "test response", recorder.Body.String())
	})

	t.Run("logs POST request with body", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created"))
		})

		handler := LoggingMiddleware(testHandler)

		body := strings.NewReader(`{"title":"Test Book","author":"Test Author"}`)
		req := httptest.NewRequest("POST", "/api/books", body)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.Equal(t, "created", recorder.Body.String())
	})

	t.Run("logs error responses", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
		})

		handler := LoggingMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/error", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "bad request", recorder.Body.String())
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("ok"))
				})

				handler := LoggingMiddleware(testHandler)

				req := httptest.NewRequest(method, "/test", nil)
				recorder := httptest.NewRecorder()

				handler.ServeHTTP(recorder, req)

				assert.Equal(t, http.StatusOK, recorder.Code)
			})
		}
	})
}

func TestCORSMiddleware(t *testing.T) {
	t.Run("adds CORS headers", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		handler := CORSMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		// Check CORS headers
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "test response", recorder.Body.String())
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This should not be called for OPTIONS requests
			t.Error("Handler should not be called for OPTIONS request")
		})

		handler := CORSMiddleware(testHandler)

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		// Check CORS headers
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))

		// OPTIONS should return 200 OK
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("adds CORS headers to error responses", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		})

		handler := CORSMiddleware(testHandler)

		req := httptest.NewRequest("POST", "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		// Check CORS headers are still present
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "error", recorder.Body.String())
	})
}

func TestJSONMiddleware(t *testing.T) {
	t.Run("sets JSON content type", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"test"}`))
		})

		handler := JSONMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `{"message":"test"}`, recorder.Body.String())
	})

	t.Run("preserves existing content type if set by handler", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("plain text"))
		})

		handler := JSONMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		// The handler sets the content type after middleware, so it should override
		assert.Equal(t, "text/plain", recorder.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "plain text", recorder.Body.String())
	})

	t.Run("works with different response codes", func(t *testing.T) {
		codes := []int{200, 201, 400, 404, 500}

		for _, code := range codes {
			t.Run(http.StatusText(code), func(t *testing.T) {
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(code)
					w.Write([]byte(`{"status":"test"}`))
				})

				handler := JSONMiddleware(testHandler)

				req := httptest.NewRequest("GET", "/test", nil)
				recorder := httptest.NewRecorder()

				handler.ServeHTTP(recorder, req)

				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
				assert.Equal(t, code, recorder.Code)
			})
		}
	})
}

func TestMiddlewareChaining(t *testing.T) {
	t.Run("chain all middleware together", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]string{"message": "success"}
			json.NewEncoder(w).Encode(response)
		})

		// Chain all middleware
		handler := LoggingMiddleware(
			CORSMiddleware(
				JSONMiddleware(testHandler),
			),
		)

		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		// Check that all middleware effects are present
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Verify JSON response
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})

	t.Run("chain handles OPTIONS request", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called for OPTIONS request")
		})

		handler := LoggingMiddleware(
			CORSMiddleware(
				JSONMiddleware(testHandler),
			),
		)

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		// CORS middleware should handle OPTIONS and not call the handler
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("chain preserves error responses", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]string{"error": "bad request"}
			json.NewEncoder(w).Encode(response)
		})

		handler := LoggingMiddleware(
			CORSMiddleware(
				JSONMiddleware(testHandler),
			),
		)

		req := httptest.NewRequest("POST", "/test", strings.NewReader("invalid"))
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "bad request", response["error"])
	})
}

func TestMiddlewareHeaders(t *testing.T) {
	t.Run("test request headers preservation", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that request headers are preserved
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		})

		handler := LoggingMiddleware(
			CORSMiddleware(
				JSONMiddleware(testHandler),
			),
		)

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test":"data"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer token123")

		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestMiddlewareWithPanic(t *testing.T) {
	t.Run("middleware should not interfere with panic recovery", func(t *testing.T) {
		// Note: This test doesn't include panic recovery middleware
		// In a real application, you'd want to add panic recovery middleware
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		handler := LoggingMiddleware(
			CORSMiddleware(
				JSONMiddleware(testHandler),
			),
		)

		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		// This should not panic
		assert.NotPanics(t, func() {
			handler.ServeHTTP(recorder, req)
		})

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	// Create a test handler that checks for request ID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("Expected request ID to be set in context")
		}

		// Check if request ID header is set in response
		responseRequestID := w.Header().Get("X-Request-ID")
		if responseRequestID == "" {
			t.Error("Expected X-Request-ID header to be set in response")
		}

		if requestID != responseRequestID {
			t.Errorf("Context request ID (%s) should match response header (%s)", requestID, responseRequestID)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with RequestIDMiddleware
	handler := RequestIDMiddleware(testHandler)

	// Test without existing request ID
	t.Run("generates new request ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		requestID := w.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("Expected X-Request-ID header to be set")
		}
	})

	// Test with existing request ID
	t.Run("uses existing request ID", func(t *testing.T) {
		existingID := "test-request-id-123"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		requestID := w.Header().Get("X-Request-ID")
		if requestID != existingID {
			t.Errorf("Expected request ID to be %s, got %s", existingID, requestID)
		}
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	// Create a test handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with RecoveryMiddleware
	handler := RecoveryMiddleware(panicHandler)

	t.Run("recovers from panic", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// This should not panic
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type to be application/json, got %s", contentType)
		}
	})
}

func TestGetRequestID(t *testing.T) {
	t.Run("returns empty string when no request ID in context", func(t *testing.T) {
		ctx := context.Background()
		requestID := GetRequestID(ctx)
		if requestID != "" {
			t.Errorf("Expected empty string, got %s", requestID)
		}
	})

	t.Run("returns request ID from context", func(t *testing.T) {
		expectedID := "test-id-123"
		ctx := context.WithValue(context.Background(), RequestIDKey, expectedID)
		requestID := GetRequestID(ctx)
		if requestID != expectedID {
			t.Errorf("Expected %s, got %s", expectedID, requestID)
		}
	})
}
