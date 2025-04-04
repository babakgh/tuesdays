package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestHandler_HandleWebSocket(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := NewHandler()
		handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + server.URL[4:]

	// Create a WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Test sending a message
	message := map[string]interface{}{
		"command": "broadcast",
		"message": "Hello, world!",
	}
	if err := conn.WriteJSON(message); err != nil {
		t.Errorf("Failed to write message: %v", err)
	}

	// Test sending list command
	listMessage := map[string]interface{}{
		"command": "list",
	}
	if err := conn.WriteJSON(listMessage); err != nil {
		t.Errorf("Failed to write list message: %v", err)
	}
} 