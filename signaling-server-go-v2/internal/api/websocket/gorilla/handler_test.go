package gorilla

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/metrics"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/tracing"
)

// MockLogger implements logging.Logger for testing
type MockLogger struct{}

func (l *MockLogger) Debug(msg string, keyvals ...interface{})   {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Error(msg string, keyvals ...interface{})   {}
func (l *MockLogger) With(keyvals ...interface{}) logging.Logger { return l }

func TestNewHandler(t *testing.T) {
	// Create config
	cfg := config.WebSocketConfig{
		Path:           "/ws",
		PingInterval:   30,
		PongWait:       60,
		WriteWait:      10,
		MaxMessageSize: 1024 * 1024,
	}

	// Create dependencies
	logger := &MockLogger{}
	metrics := metrics.NewMetrics(config.MetricsConfig{Enabled: true})
	tracer := &tracing.NoopTracer{}

	// Create handler
	handler := NewHandler(cfg, logger, metrics, tracer)

	// Verify handler is not nil
	if handler == nil {
		t.Fatal("Handler should not be nil")
	}
}

func TestHandleConnection(t *testing.T) {
	// Create config
	cfg := config.WebSocketConfig{
		Path:           "/ws",
		PingInterval:   30,
		PongWait:       60,
		WriteWait:      10,
		MaxMessageSize: 1024 * 1024,
	}

	// Create dependencies
	logger := &MockLogger{}
	metrics := metrics.NewMetrics(config.MetricsConfig{Enabled: true})
	tracer := &tracing.NoopTracer{}

	// Create handler
	handler := NewHandler(cfg, logger, metrics, tracer)

	// Create a test request
	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.HandleConnection(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify content type
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestBroadcastAndSendMessages(t *testing.T) {
	// Create handler
	cfg := config.WebSocketConfig{
		Path:           "/ws",
		PingInterval:   30,
		PongWait:       60,
		WriteWait:      10,
		MaxMessageSize: 1024 * 1024,
	}
	logger := &MockLogger{}
	metrics := metrics.NewMetrics(config.MetricsConfig{Enabled: true})
	tracer := &tracing.NoopTracer{}
	h := NewHandler(cfg, logger, metrics, tracer).(*Handler)

	// Manually create and add a test client
	h.mux.Lock()
	clientID := "test-client"
	client := &Client{
		id:      clientID,
		handler: h,
		send:    make(chan []byte, 10),
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
	h.clients[clientID] = client
	h.mux.Unlock()

	// Test sending a message
	message := []byte("test message")
	err := h.SendMessage(clientID, message)
	if err != nil {
		t.Errorf("SendMessage failed: %v", err)
	}

	// Verify message was sent
	receivedMsg := <-client.send
	if string(receivedMsg) != string(message) {
		t.Errorf("Expected message %s, got %s", string(message), string(receivedMsg))
	}

	// Test broadcasting a message
	broadcastMsg := []byte("broadcast test")
	err = h.BroadcastMessage(broadcastMsg)
	if err != nil {
		t.Errorf("BroadcastMessage failed: %v", err)
	}

	// Verify broadcast message was received
	broadcastReceived := <-client.send
	if string(broadcastReceived) != string(broadcastMsg) {
		t.Errorf("Expected broadcast message %s, got %s", string(broadcastMsg), string(broadcastReceived))
	}

	// Test closing a connection
	err = h.CloseConnection(clientID)
	if err != nil {
		t.Errorf("CloseConnection failed: %v", err)
	}

	// Verify client was removed
	h.mux.Lock()
	_, exists := h.clients[clientID]
	h.mux.Unlock()

	if exists {
		t.Error("Expected client to be removed after closing connection")
	}
}
