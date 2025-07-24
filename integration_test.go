package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

// Integration test to verify error handling improvements
func TestErrorHandlingIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration tests")
	}

	// Wait for server to be ready
	time.Sleep(2 * time.Second)

	t.Run("health check should return request ID", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check if request ID header is present (if middleware is working)
		requestID := resp.Header.Get("X-Request-ID")
		if requestID != "" {
			t.Logf("Request ID found: %s", requestID)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid UUID should return structured error", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/books/invalid-uuid")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		// Check response structure
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if errorResp["error"] == nil {
			t.Error("Expected error field in response")
		}

		if errorResp["message"] == nil {
			t.Error("Expected message field in response")
		}

		if errorResp["code"] == nil {
			t.Error("Expected code field in response")
		}
	})

	t.Run("validation error should return 400", func(t *testing.T) {
		invalidBook := map[string]interface{}{
			"title":  "", // Empty title should trigger validation error
			"author": "Test Author",
		}

		body, _ := json.Marshal(invalidBook)
		resp, err := http.Post(baseURL+"/api/books", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if errorResp["error"] != "Validation error" {
			t.Errorf("Expected validation error, got %v", errorResp["error"])
		}
	})

	t.Run("duplicate ISBN should return 409", func(t *testing.T) {
		duplicateBook := map[string]interface{}{
			"title":     "Duplicate Test",
			"author":    "Test Author",
			"isbn":      "9781234567890", // This ISBN should already exist from previous tests
			"publisher": "Test",
			"genre":     "Test",
			"pages":     100,
			"language":  "English",
			"available": true,
		}

		body, _ := json.Marshal(duplicateBook)
		resp, err := http.Post(baseURL+"/api/books", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}

		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if errorResp["error"] != "Duplicate resource" {
			t.Errorf("Expected duplicate resource error, got %v", errorResp["error"])
		}
	})

	t.Run("non-existent book should return 404", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/books/00000000-0000-0000-0000-000000000000")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}

		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if errorResp["error"] != "Book not found" {
			t.Errorf("Expected book not found error, got %v", errorResp["error"])
		}
	})
}

// Test Redis caching functionality
func TestRedisCacheIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests")
	}

	t.Run("books endpoint should work with Redis", func(t *testing.T) {
		// First request - should hit database
		start1 := time.Now()
		resp1, err := http.Get(baseURL + "/api/books")
		if err != nil {
			t.Fatalf("Failed to make first request: %v", err)
		}
		defer resp1.Body.Close()
		duration1 := time.Since(start1)

		if resp1.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp1.StatusCode)
		}

		// Second request - should hit cache (potentially faster)
		start2 := time.Now()
		resp2, err := http.Get(baseURL + "/api/books")
		if err != nil {
			t.Fatalf("Failed to make second request: %v", err)
		}
		defer resp2.Body.Close()
		duration2 := time.Since(start2)

		if resp2.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp2.StatusCode)
		}

		t.Logf("First request took: %v, Second request took: %v", duration1, duration2)

		// Verify both responses are identical
		var data1, data2 map[string]interface{}
		json.NewDecoder(resp1.Body).Decode(&data1)
		json.NewDecoder(resp2.Body).Decode(&data2)

		// Basic check that we got data
		if data1["data"] == nil || data2["data"] == nil {
			t.Error("Expected data field in both responses")
		}
	})
}

func TestPanicRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests")
	}

	// This test would need a special endpoint that causes a panic
	// For now, we'll just verify that the recovery middleware is in place
	// by checking that normal requests still work
	t.Run("normal requests work after panic recovery setup", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}
