package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppError(t *testing.T) {
	t.Run("creates validation error", func(t *testing.T) {
		err := Validation("Invalid input", "Name is required")

		if err.Code != CodeValidation {
			t.Errorf("Expected code %s, got %s", CodeValidation, err.Code)
		}

		if err.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
		}

		if err.Message != "Invalid input" {
			t.Errorf("Expected message 'Invalid input', got %s", err.Message)
		}

		if err.Details != "Name is required" {
			t.Errorf("Expected details 'Name is required', got %s", err.Details)
		}
	})

	t.Run("creates not found error", func(t *testing.T) {
		err := NotFound("Book")

		if err.Code != CodeNotFound {
			t.Errorf("Expected code %s, got %s", CodeNotFound, err.Code)
		}

		if err.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, err.StatusCode)
		}

		if err.Message != "Book not found" {
			t.Errorf("Expected message 'Book not found', got %s", err.Message)
		}
	})

	t.Run("creates conflict error", func(t *testing.T) {
		err := Conflict("Duplicate entry", "ISBN already exists")

		if err.Code != CodeConflict {
			t.Errorf("Expected code %s, got %s", CodeConflict, err.Code)
		}

		if err.StatusCode != http.StatusConflict {
			t.Errorf("Expected status code %d, got %d", http.StatusConflict, err.StatusCode)
		}
	})

	t.Run("creates internal error", func(t *testing.T) {
		err := Internal("Database error", "Connection failed")

		if err.Code != CodeInternal {
			t.Errorf("Expected code %s, got %s", CodeInternal, err.Code)
		}

		if err.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, err.StatusCode)
		}
	})
}

func TestWriteErrorResponse(t *testing.T) {
	t.Run("writes app error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := Validation("Invalid input", "Name is required")
		requestID := "test-request-123"

		WriteErrorResponse(w, err, requestID)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type to be application/json, got %s", contentType)
		}

		body := w.Body.String()
		if body == "" {
			t.Error("Expected response body to be non-empty")
		}

		// Check if request ID is included in response
		if !containsString(body, requestID) {
			t.Errorf("Expected response to contain request ID %s", requestID)
		}
	})

	t.Run("writes generic error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := &TestError{message: "generic error"}
		requestID := "test-request-456"

		WriteErrorResponse(w, err, requestID)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestErrorCodes(t *testing.T) {
	testCases := []struct {
		code       ErrorCode
		statusCode int
	}{
		{CodeValidation, http.StatusBadRequest},
		{CodeNotFound, http.StatusNotFound},
		{CodeConflict, http.StatusConflict},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeRateLimit, http.StatusTooManyRequests},
		{CodeTimeout, http.StatusRequestTimeout},
		{CodeInternal, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(string(tc.code), func(t *testing.T) {
			statusCode := getStatusCodeForErrorCode(tc.code)
			if statusCode != tc.statusCode {
				t.Errorf("Expected status code %d for %s, got %d", tc.statusCode, tc.code, statusCode)
			}
		})
	}
}

// Helper types and functions for testing
type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
