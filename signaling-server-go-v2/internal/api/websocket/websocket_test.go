package websocket

import (
	"testing"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
)

func TestNewWebSocketConfig(t *testing.T) {
	// Create a config
	cfgInput := config.WebSocketConfig{
		Path:           "/ws",
		PingInterval:   30,
		PongWait:       60,
		WriteWait:      10,
		MaxMessageSize: 1024 * 1024,
	}

	// Convert to WebSocketConfig
	wsConfig := NewWebSocketConfig(cfgInput)

	// Verify values
	if wsConfig.Path != "/ws" {
		t.Errorf("Expected path /ws, got %s", wsConfig.Path)
	}

	expectedPingInterval := 30 * time.Second
	if wsConfig.PingInterval != expectedPingInterval {
		t.Errorf("Expected ping interval %v, got %v", expectedPingInterval, wsConfig.PingInterval)
	}

	expectedPongWait := 60 * time.Second
	if wsConfig.PongWait != expectedPongWait {
		t.Errorf("Expected pong wait %v, got %v", expectedPongWait, wsConfig.PongWait)
	}

	expectedWriteWait := 10 * time.Second
	if wsConfig.WriteWait != expectedWriteWait {
		t.Errorf("Expected write wait %v, got %v", expectedWriteWait, wsConfig.WriteWait)
	}

	expectedMaxMessageSize := int64(1024 * 1024)
	if wsConfig.MaxMessageSize != expectedMaxMessageSize {
		t.Errorf("Expected max message size %d, got %d", expectedMaxMessageSize, wsConfig.MaxMessageSize)
	}
}
