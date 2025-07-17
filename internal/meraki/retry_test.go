package meraki

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetryLogic(t *testing.T) {
	t.Run("successful request on first try", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(`{"success": true}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		client.baseURL = server.URL

		resp, err := client.makeRequest("GET", "/test")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if resp == nil {
			t.Error("Expected response, got nil")
		}
		if resp != nil {
			resp.Body.Close()
		}
	})

	t.Run("retry on 429 rate limit", func(t *testing.T) {
		attemptCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount <= 2 {
				w.WriteHeader(429) // Too Many Requests
				return
			}
			w.WriteHeader(200)
			w.Write([]byte(`{"success": true}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		client.baseURL = server.URL

		// Set short retry intervals for testing
		client.SetRetryConfig(RetryConfig{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		})

		resp, err := client.makeRequest("GET", "/test")
		if err != nil {
			t.Errorf("Expected no error after retries, got: %v", err)
		}
		if resp == nil {
			t.Error("Expected response, got nil")
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attemptCount != 3 {
			t.Errorf("Expected 3 attempts, got %d", attemptCount)
		}
	})

	t.Run("retry on 500 server error", func(t *testing.T) {
		attemptCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount <= 1 {
				w.WriteHeader(500) // Internal Server Error
				return
			}
			w.WriteHeader(200)
			w.Write([]byte(`{"success": true}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		client.baseURL = server.URL

		// Set short retry intervals for testing
		client.SetRetryConfig(RetryConfig{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		})

		resp, err := client.makeRequest("GET", "/test")
		if err != nil {
			t.Errorf("Expected no error after retries, got: %v", err)
		}
		if resp == nil {
			t.Error("Expected response, got nil")
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attemptCount != 2 {
			t.Errorf("Expected 2 attempts, got %d", attemptCount)
		}
	})

	t.Run("no retry on 404 not found", func(t *testing.T) {
		attemptCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.WriteHeader(404) // Not Found
		}))
		defer server.Close()

		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		client.baseURL = server.URL

		resp, err := client.makeRequest("GET", "/test")
		if err == nil {
			t.Error("Expected error for 404, got none")
		}
		if resp != nil {
			t.Error("Expected no response for 404, got one")
		}
		if attemptCount != 1 {
			t.Errorf("Expected 1 attempt (no retry), got %d", attemptCount)
		}
	})

	t.Run("exhaust all retries", func(t *testing.T) {
		attemptCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.WriteHeader(500) // Always return server error
		}))
		defer server.Close()

		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		client.baseURL = server.URL

		// Set short retry intervals for testing
		client.SetRetryConfig(RetryConfig{
			MaxRetries:      2,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		})

		resp, err := client.makeRequest("GET", "/test")
		if err == nil {
			t.Error("Expected error after exhausting retries, got none")
		}
		if resp != nil {
			t.Error("Expected no response after exhausting retries, got one")
		}
		if attemptCount != 3 { // MaxRetries + 1
			t.Errorf("Expected 3 attempts, got %d", attemptCount)
		}
	})
}

func TestRetryConfig(t *testing.T) {
	t.Run("default retry config", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		config := client.GetRetryConfig()
		if config.MaxRetries != 3 {
			t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
		}
		if config.InitialInterval != 1*time.Second {
			t.Errorf("Expected InitialInterval=1s, got %v", config.InitialInterval)
		}
		if config.MaxInterval != 30*time.Second {
			t.Errorf("Expected MaxInterval=30s, got %v", config.MaxInterval)
		}
		if config.Multiplier != 2.0 {
			t.Errorf("Expected Multiplier=2.0, got %f", config.Multiplier)
		}
	})

	t.Run("custom retry config", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		customConfig := RetryConfig{
			MaxRetries:      5,
			InitialInterval: 500 * time.Millisecond,
			MaxInterval:     60 * time.Second,
			Multiplier:      3.0,
		}

		client.SetRetryConfig(customConfig)
		config := client.GetRetryConfig()

		if config.MaxRetries != 5 {
			t.Errorf("Expected MaxRetries=5, got %d", config.MaxRetries)
		}
		if config.InitialInterval != 500*time.Millisecond {
			t.Errorf("Expected InitialInterval=500ms, got %v", config.InitialInterval)
		}
		if config.MaxInterval != 60*time.Second {
			t.Errorf("Expected MaxInterval=60s, got %v", config.MaxInterval)
		}
		if config.Multiplier != 3.0 {
			t.Errorf("Expected Multiplier=3.0, got %f", config.Multiplier)
		}
	})
}

func TestCalculateBackoff(t *testing.T) {
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test exponential backoff
	backoff0 := client.calculateBackoff(0)
	backoff1 := client.calculateBackoff(1)
	backoff2 := client.calculateBackoff(2)

	if backoff0 != 1*time.Second {
		t.Errorf("Expected backoff0=1s, got %v", backoff0)
	}
	if backoff1 != 1*time.Second {
		t.Errorf("Expected backoff1=1s, got %v", backoff1)
	}
	if backoff2 != 2*time.Second {
		t.Errorf("Expected backoff2=2s, got %v", backoff2)
	}

	// Test maximum backoff cap
	client.SetRetryConfig(RetryConfig{
		MaxRetries:      10,
		InitialInterval: 1 * time.Second,
		MaxInterval:     5 * time.Second,
		Multiplier:      2.0,
	})

	backoff10 := client.calculateBackoff(10)
	if backoff10 != 5*time.Second {
		t.Errorf("Expected backoff capped at 5s, got %v", backoff10)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		want       bool
	}{
		{"network error", http.ErrHandlerTimeout, 0, true},
		{"429 rate limit", nil, 429, true},
		{"500 server error", nil, 500, true},
		{"502 bad gateway", nil, 502, true},
		{"503 service unavailable", nil, 503, true},
		{"504 gateway timeout", nil, 504, true},
		{"400 bad request", nil, 400, false},
		{"401 unauthorized", nil, 401, false},
		{"403 forbidden", nil, 403, false},
		{"404 not found", nil, 404, false},
		{"200 success", nil, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableError(tt.err, tt.statusCode); got != tt.want {
				t.Errorf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}
