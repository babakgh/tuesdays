package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/metrics"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/tracing"
)

// MockLogger implements logging.Logger for testing
type MockLogger struct {
	DebugCalled bool
	InfoCalled  bool
	WarnCalled  bool
	ErrorCalled bool
}

func (l *MockLogger) Debug(msg string, keyvals ...interface{})   { l.DebugCalled = true }
func (l *MockLogger) Info(msg string, keyvals ...interface{})    { l.InfoCalled = true }
func (l *MockLogger) Warn(msg string, keyvals ...interface{})    { l.WarnCalled = true }
func (l *MockLogger) Error(msg string, keyvals ...interface{})   { l.ErrorCalled = true }
func (l *MockLogger) With(keyvals ...interface{}) logging.Logger { return l }

// Test Logging middleware
func TestLoggingMiddleware(t *testing.T) {
	mockLogger := &MockLogger{}

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply middleware
	handler := Logging(mockLogger)(nextHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	rec := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rec, req)

	// Verify the logger was called
	if !mockLogger.InfoCalled {
		t.Error("Expected logger.Info to be called")
	}

	// Verify the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

// Test Recovery middleware
func TestRecoveryMiddleware(t *testing.T) {
	mockLogger := &MockLogger{}

	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Apply recovery middleware
	handler := Recovery(mockLogger)(panicHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	rec := httptest.NewRecorder()

	// This should not panic
	handler.ServeHTTP(rec, req)

	// Verify the logger was called with error
	if !mockLogger.ErrorCalled {
		t.Error("Expected logger.Error to be called for panic")
	}

	// Verify proper status code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// Test Metrics middleware
func TestMetricsMiddleware(t *testing.T) {
	// Create config for metrics
	metricsCfg := config.MetricsConfig{
		Enabled: true,
		Path:    "/metrics",
	}

	// Create actual metrics instance
	mockMetrics := metrics.NewMetrics(metricsCfg)

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("metrics test"))
	})

	// Apply middleware
	handler := Metrics(mockMetrics)(nextHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/metrics-test", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rec, req)

	// We can't verify metrics were recorded since we're using the real metrics object
	// which doesn't provide recording status feedback. In a real test, we could
	// use a mock implementation with verification.

	// Verify the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

// MockTracer implements tracing.Tracer for testing
type MockTracer struct{}

func (t *MockTracer) StartSpan(name string, opts ...tracing.SpanOption) tracing.Span {
	return &MockSpan{}
}

func (t *MockTracer) Inject(ctx context.Context, carrier interface{}) error {
	return nil
}

func (t *MockTracer) Extract(carrier interface{}) (context.Context, error) {
	return context.Background(), nil
}

// MockSpan implements tracing.Span for testing
type MockSpan struct {
	EndCalled  bool
	Attributes map[string]interface{}
}

func (s *MockSpan) End() {
	s.EndCalled = true
}

func (s *MockSpan) SetAttribute(key string, value interface{}) {
	if s.Attributes == nil {
		s.Attributes = make(map[string]interface{})
	}
	s.Attributes[key] = value
}

func (s *MockSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *MockSpan) RecordError(err error)                                   {}
func (s *MockSpan) Context() context.Context                                { return context.Background() }

// Test Tracing middleware
func TestTracingMiddleware(t *testing.T) {
	mockTracer := &MockTracer{}

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("tracing test"))
	})

	// Apply middleware
	handler := Tracing(mockTracer)(nextHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/tracing-test", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rec, req)

	// Verify the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}
