package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
)

// MockLogger implements logging.Logger for testing
type MockLogger struct{}

func (l *MockLogger) Debug(msg string, keyvals ...interface{})   {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Error(msg string, keyvals ...interface{})   {}
func (l *MockLogger) With(keyvals ...interface{}) logging.Logger { return l }

func TestLivenessHandler(t *testing.T) {
	// Create a new health handler
	handler := NewHandler(&MockLogger{})

	// Add a check that will pass
	handler.AddLivenessCheck("service-status", func() (Status, string) {
		return StatusUp, "Service is running"
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/health/live", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.LiveHandler(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify the response body
	var response HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Status != StatusUp {
		t.Errorf("Expected status %s, got %s", StatusUp, response.Status)
	}

	if check, ok := response.Checks["service-status"]; !ok {
		t.Error("Expected 'service-status' check in response")
	} else if check.Status != StatusUp {
		t.Errorf("Expected check status %s, got %s", StatusUp, check.Status)
	}
}

func TestLivenessHandlerWithFailedCheck(t *testing.T) {
	// Create a new health handler
	handler := NewHandler(&MockLogger{})

	// Add a check that will fail
	handler.AddLivenessCheck("failing-check", func() (Status, string) {
		return StatusDown, "Service is down"
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/health/live", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.LiveHandler(rec, req)

	// Check the response
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	// Verify the response body
	var response HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Status != StatusDown {
		t.Errorf("Expected status %s, got %s", StatusDown, response.Status)
	}
}

func TestReadinessHandler(t *testing.T) {
	// Create a new health handler
	handler := NewHandler(&MockLogger{})

	// Add a check that will pass
	handler.AddReadinessCheck("database-connection", func() (Status, string) {
		return StatusUp, "Database is connected"
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/health/ready", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.ReadyHandler(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify the response body
	var response HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Status != StatusUp {
		t.Errorf("Expected status %s, got %s", StatusUp, response.Status)
	}

	if check, ok := response.Checks["database-connection"]; !ok {
		t.Error("Expected 'database-connection' check in response")
	} else if check.Status != StatusUp {
		t.Errorf("Expected check status %s, got %s", StatusUp, check.Status)
	}
}

func TestReadinessHandlerWithFailedCheck(t *testing.T) {
	// Create a new health handler
	handler := NewHandler(&MockLogger{})

	// Add a check that will fail
	handler.AddReadinessCheck("external-api", func() (Status, string) {
		return StatusDown, "External API is not responding"
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/health/ready", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.ReadyHandler(rec, req)

	// Check the response
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	// Verify the response body
	var response HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Status != StatusDown {
		t.Errorf("Expected status %s, got %s", StatusDown, response.Status)
	}
}
