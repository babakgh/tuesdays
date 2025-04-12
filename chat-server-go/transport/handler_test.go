package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	
	// Test sending dm command with missing recipient
	dmMessageWithoutRecipient := map[string]interface{}{
		"command": "dm",
		"message": "Hello, direct message!",
	}
	if err := conn.WriteJSON(dmMessageWithoutRecipient); err != nil {
		t.Errorf("Failed to write DM message: %v", err)
	}
	
	// Create a second connection to test DM functionality
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect second WebSocket: %v", err)
	}
	defer conn2.Close()
	
	// Allow time for the second connection to be established
	time.Sleep(100 * time.Millisecond)
	
	// Just use a simple hardcoded recipient name
	// We know the handler generates names like "member1", "member2", etc.
	// The first connection is "member1", so the second will be "member2"
	recipient := "member2"
	
	// Allow time for the connections to be registered
	time.Sleep(100 * time.Millisecond)
	
	// Test sending a valid DM
	validDM := map[string]interface{}{
		"command":   "dm",
		"message":   "Hello, direct message!",
		"recipient": recipient,
	}
	if err := conn.WriteJSON(validDM); err != nil {
		t.Errorf("Failed to write valid DM: %v", err)
	}
} 