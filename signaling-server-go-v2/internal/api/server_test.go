package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/metrics"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/tracing"
)

// MockRouter implements the Router interface for testing
type MockRouter struct {
	handlers map[string]http.Handler
	mws      []func(http.Handler) http.Handler
}

func NewMockRouter() *MockRouter {
	return &MockRouter{
		handlers: make(map[string]http.Handler),
		mws:      []func(http.Handler) http.Handler{},
	}
}

func (r *MockRouter) Handle(method, path string, handler http.Handler) {
	r.handlers[method+":"+path] = handler
}

func (r *MockRouter) HandleFunc(method, path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.handlers[method+":"+path] = http.HandlerFunc(handlerFunc)
}

func (r *MockRouter) Use(middleware ...func(http.Handler) http.Handler) {
	r.mws = append(r.mws, middleware...)
}

func (r *MockRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.Method + ":" + req.URL.Path
	if handler, ok := r.handlers[key]; ok {
		handler.ServeHTTP(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// MockLogger implements logging.Logger for testing
type MockLogger struct{}

func (l *MockLogger) Debug(msg string, keyvals ...interface{})   {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Error(msg string, keyvals ...interface{})   {}
func (l *MockLogger) With(keyvals ...interface{}) logging.Logger { return l }

// MockWebSocketHandler implements the WebSocketHandler interface for testing
type MockWebSocketHandler struct{}

func (h *MockWebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("WebSocket connection"))
}

func (h *MockWebSocketHandler) BroadcastMessage(message []byte) error {
	return nil
}

func (h *MockWebSocketHandler) SendMessage(clientID string, message []byte) error {
	return nil
}

func (h *MockWebSocketHandler) CloseConnection(clientID string) error {
	return nil
}

func setupTestServer() (*Server, *MockRouter) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:            8080,
			Host:            "127.0.0.1",
			ShutdownTimeout: 1, // Short timeout for testing
			ReadTimeout:     1,
			WriteTimeout:    1,
			IdleTimeout:     1,
		},
		Monitoring: config.MonitoringConfig{
			LivenessPath:  "/health/live",
			ReadinessPath: "/health/ready",
		},
		WebSocket: config.WebSocketConfig{
			Path: "/ws",
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}

	router := NewMockRouter()
	logger := &MockLogger{}
	m := metrics.NewMetrics(cfg.Metrics)
	tracer := &tracing.NoopTracer{}
	wsHandler := &MockWebSocketHandler{}

	server := NewServer(cfg, router, logger, m, tracer, wsHandler)
	return server, router
}

func TestServerRouteRegistration(t *testing.T) {
	server, mockRouter := setupTestServer()

	// Check that health endpoints are registered
	if _, ok := mockRouter.handlers["GET:/health/live"]; !ok {
		t.Error("Liveness endpoint not registered")
	}

	if _, ok := mockRouter.handlers["GET:/health/ready"]; !ok {
		t.Error("Readiness endpoint not registered")
	}

	// Check that WebSocket endpoint is registered
	if _, ok := mockRouter.handlers["GET:/ws"]; !ok {
		t.Error("WebSocket endpoint not registered")
	}

	// Make sure server was created with expected address
	expectedAddr := "127.0.0.1:8080"
	if server.httpServer.Addr != expectedAddr {
		t.Errorf("Expected server address %s, got %s", expectedAddr, server.httpServer.Addr)
	}
}

func TestHealthEndpoints(t *testing.T) {
	_, mockRouter := setupTestServer()

	// Test liveness endpoint
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/live", nil)
	mockRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Test readiness endpoint
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/health/ready", nil)
	mockRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestServerShutdown(t *testing.T) {
	server, _ := setupTestServer()

	// Start server in a goroutine
	go func() {
		// Start for a short time, then exit - we're not actually binding to a port in tests
		time.Sleep(100 * time.Millisecond)
	}()

	// Try graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should not block or error since we're not really starting the server
	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Server shutdown failed: %v", err)
	}
}
