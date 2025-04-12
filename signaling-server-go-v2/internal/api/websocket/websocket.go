package websocket

import (
	"net/http"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
)

// WebSocketHandler interface for abstracting WebSocket implementations
type WebSocketHandler interface {
	HandleConnection(w http.ResponseWriter, r *http.Request)
	BroadcastMessage(message []byte) error
	SendMessage(clientID string, message []byte) error
	CloseConnection(clientID string) error
}

// WebSocketConnection interface for abstracting WebSocket connection implementations
type WebSocketConnection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

// WebSocketConfig holds configuration for WebSocket connections
type WebSocketConfig struct {
	Path           string
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteWait      time.Duration
	MaxMessageSize int64
}

// NewWebSocketConfig creates a WebSocketConfig from config.WebSocketConfig
func NewWebSocketConfig(cfg config.WebSocketConfig) WebSocketConfig {
	return WebSocketConfig{
		Path:           cfg.Path,
		PingInterval:   time.Duration(cfg.PingInterval) * time.Second,
		PongWait:       time.Duration(cfg.PongWait) * time.Second,
		WriteWait:      time.Duration(cfg.WriteWait) * time.Second,
		MaxMessageSize: cfg.MaxMessageSize,
	}
}
