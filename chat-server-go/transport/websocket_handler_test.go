package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chat-server-go/domain"
	"github.com/gorilla/websocket"
)

func TestWebSocketHandler_HandleWebSocket(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := NewWebSocketHandler()
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

func TestWebSocketHandler_sendWelcomeMessages(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "1",
		Name: "test",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Test sendWelcomeMessages
	handler.sendWelcomeMessages(member)

	// Verify me event was sent
	select {
	case msg := <-mockConn.writeChan:
		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("Failed to unmarshal me event: %v", err)
		}
		if event["event"] != "me" {
			t.Errorf("Expected event type 'me', got %v", event["event"])
		}
	default:
		t.Error("No me event was sent")
	}

	// Verify join broadcast was sent
	select {
	case msg := <-mockConn.writeChan:
		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("Failed to unmarshal join event: %v", err)
		}
		if event["event"] != "broadcast" {
			t.Errorf("Expected event type 'broadcast', got %v", event["event"])
		}
	default:
		t.Error("No join broadcast was sent")
	}
}

func TestWebSocketHandler_handleMessages(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "1",
		Name: "test",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Start message handling in a goroutine
	go handler.handleMessages(member)

	// Test broadcast message
	broadcastMsg := map[string]interface{}{
		"command": "broadcast",
		"message": "Hello, world!",
	}
	data, _ := json.Marshal(broadcastMsg)
	mockConn.readChan <- data

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify broadcast was sent
	select {
	case msg := <-mockConn.writeChan:
		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("Failed to unmarshal broadcast event: %v", err)
		}
		if event["event"] != "broadcast" {
			t.Errorf("Expected event type 'broadcast', got %v", event["event"])
		}
	default:
		t.Error("No broadcast was sent")
	}

	// Test list command
	listMsg := map[string]interface{}{
		"command": "list",
	}
	data, _ = json.Marshal(listMsg)
	mockConn.readChan <- data

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify list response was sent
	select {
	case msg := <-mockConn.writeChan:
		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("Failed to unmarshal list event: %v", err)
		}
		if event["event"] != "list" {
			t.Errorf("Expected event type 'list', got %v", event["event"])
		}
	default:
		t.Error("No list response was sent")
	}
	
	// Test DM command (with invalid recipient)
	invalidDmMsg := map[string]interface{}{
		"command":   "dm",
		"message":   "Hello, this is a direct message!",
		"recipient": "nonexistent-member", // This member doesn't exist
	}
	data, _ = json.Marshal(invalidDmMsg)
	mockConn.readChan <- data
	
	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)
	
	// Verify error response was sent for invalid recipient
	select {
	case msg := <-mockConn.writeChan:
		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("Failed to unmarshal error event: %v", err)
		}
		if event["event"] != "error" {
			t.Errorf("Expected event type 'error', got %v", event["event"])
		}
	default:
		t.Error("No error response was sent for invalid recipient")
	}
	
	// Create another member for DM test
	mockConn2 := newMockWebSocketConn().(*mockWebSocketConn)
	member2 := &domain.Member{
		ID:   "2",
		Name: "recipient-member",
		Conn: mockConn2,
	}
	
	// Add second member to store
	if err := handler.store.Add(member2); err != nil {
		t.Fatalf("Failed to add second member: %v", err)
	}
	
	// Start message handling for second member
	go handler.handleMessages(member2)
	
	// Test valid DM command
	validDmMsg := map[string]interface{}{
		"command":   "dm",
		"message":   "Hello, this is a direct message!",
		"recipient": "recipient-member", // This member exists
	}
	data, _ = json.Marshal(validDmMsg)
	mockConn.readChan <- data
	
	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)
	
	// Verify DM was delivered to recipient
	select {
	case msg := <-mockConn2.writeChan:
		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("Failed to unmarshal dm event: %v", err)
		}
		if event["event"] != "dm" {
			t.Errorf("Expected event type 'dm', got %v", event["event"])
		}
		if event["member"] != "test" {
			t.Errorf("Expected sender 'test', got %v", event["member"])
		}
		if event["message"] != "Hello, this is a direct message!" {
			t.Errorf("Expected correct message, got %v", event["message"])
		}
	default:
		t.Error("No DM was delivered to recipient")
	}
	
	// Verify confirmation was sent to sender
	select {
	case msg := <-mockConn.writeChan:
		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("Failed to unmarshal confirmation event: %v", err)
		}
		if event["event"] != "dm_sent" {
			t.Errorf("Expected event type 'dm_sent', got %v", event["event"])
		}
	default:
		t.Error("No confirmation was sent to sender")
	}
	
	// Clean up second member
	mockConn2.Close()

	// Close the connection
	mockConn.Close()
}

func TestWebSocketHandler_HandleWebSocket_ErrorCases(t *testing.T) {
	// Test invalid WebSocket upgrade
	t.Run("invalid upgrade", func(t *testing.T) {
		handler := NewWebSocketHandler()
		req := httptest.NewRequest("GET", "/ws", nil)
		w := httptest.NewRecorder()
		handler.HandleWebSocket(w, req)
	})

	// Test member store error
	t.Run("store error", func(t *testing.T) {
		handler := NewWebSocketHandler()
		mockConn := newMockWebSocketConn().(*mockWebSocketConn)
		member := &domain.Member{
			ID:   "1",
			Name: "test",
			Conn: mockConn,
		}

		// Add member first time
		if err := handler.store.Add(member); err != nil {
			t.Fatalf("Failed to add member: %v", err)
		}

		// Try to add same member again (should fail)
		if err := handler.store.Add(member); err == nil {
			t.Error("Expected error when adding duplicate member")
		}
	})
}

func TestWebSocketHandler_handleMessages_ErrorCases(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "1",
		Name: "test",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Start message handling in a goroutine
	go handler.handleMessages(member)

	// Test invalid JSON message
	invalidJSON := []byte("invalid json")
	mockConn.readChan <- invalidJSON

	// Test invalid command
	invalidCmd := map[string]interface{}{
		"command": "invalid",
		"message": "test",
	}
	data, _ := json.Marshal(invalidCmd)
	mockConn.readChan <- data

	// Test connection close
	mockConn.Close()

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify member was removed from store
	members := handler.store.List()
	if len(members) != 0 {
		t.Errorf("Expected 0 members after disconnect, got %d", len(members))
	}
}

func TestWebSocketHandler_sendWelcomeMessages_ErrorCases(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "1",
		Name: "test",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Close connection before sending welcome messages
	mockConn.Close()

	// Test sendWelcomeMessages with closed connection
	handler.sendWelcomeMessages(member)
}

func TestWebSocketHandler_HandleWebSocket_ConnectionError(t *testing.T) {
	// Create a test server that will fail the WebSocket upgrade
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't upgrade the connection
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	// Try to connect to the server
	_, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], nil)
	if err == nil {
		t.Error("Expected connection error")
	}
}

func TestWebSocketHandler_handleMessages_CommandError(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "1",
		Name: "test",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Start message handling in a goroutine
	go handler.handleMessages(member)

	// Test command with missing required fields
	invalidCmd := map[string]interface{}{
		"command": "broadcast",
		// Missing message field
	}
	data, _ := json.Marshal(invalidCmd)
	mockConn.readChan <- data

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Test unexpected close error
	mockConn.Close()
	time.Sleep(100 * time.Millisecond)
}

// errorMockConn is a mock that always returns errors for writes
type errorMockConn struct {
	writeCount int
}

func (m *errorMockConn) ReadMessage() (messageType int, p []byte, err error) {
	return 0, nil, websocket.ErrCloseSent
}

func (m *errorMockConn) WriteJSON(v interface{}) error {
	m.writeCount++
	return websocket.ErrCloseSent
}

func (m *errorMockConn) Close() error {
	return nil
}

func TestWebSocketHandler_sendWelcomeMessages_WriteError(t *testing.T) {
	handler := NewWebSocketHandler()
	errorMock := &errorMockConn{}
	
	member := &domain.Member{
		ID:   "test-error",
		Name: "test-error",
		Conn: errorMock,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Test sendWelcomeMessages with write error
	handler.sendWelcomeMessages(member)

	// Verify that we attempted to write messages
	if errorMock.writeCount == 0 {
		t.Error("Expected write attempts, got none")
	}
}

func TestWebSocketHandler_handleMessages_ParseError(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "test-parse",
		Name: "test-parse",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Start message handling in a goroutine
	go handler.handleMessages(member)

	// Send invalid JSON
	mockConn.readChan <- []byte("{invalid json")

	// Send valid JSON with invalid command
	mockConn.readChan <- []byte(`{"command": "invalid_command"}`)

	// Wait for messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Close connection
	mockConn.Close()
}

func TestWebSocketHandler_handleMessages_CommandExecutionError(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "test-exec",
		Name: "test-exec",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Start message handling in a goroutine
	go handler.handleMessages(member)

	// Send broadcast command without message
	mockConn.readChan <- []byte(`{"command": "broadcast"}`)

	// Send list command with invalid data
	mockConn.readChan <- []byte(`{"command": "list", "data": {"invalid": true}}`)

	// Wait for messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Close connection
	mockConn.Close()
}

func TestWebSocketHandler_handleMessages_UnexpectedClose(t *testing.T) {
	handler := NewWebSocketHandler()
	mockConn := newMockWebSocketConn().(*mockWebSocketConn)
	member := &domain.Member{
		ID:   "test-close",
		Name: "test-close",
		Conn: mockConn,
	}

	// Add member to store
	if err := handler.store.Add(member); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Start message handling in a goroutine
	go handler.handleMessages(member)

	// Close the connection unexpectedly
	mockConn.Close()

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify member was removed
	if _, err := handler.store.Get(member.ID); err == nil {
		t.Error("Expected member to be removed after connection close")
	}
}

func TestWebSocketHandler_HandleWebSocket_UpgradeError(t *testing.T) {
	handler := NewWebSocketHandler()

	// Create a request without WebSocket headers
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	// This should fail because the request doesn't have WebSocket headers
	handler.HandleWebSocket(w, req)

	// Verify response indicates error
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// errorStore is a mock store that always returns an error
type errorStore struct {
	domain.MemberStore
}

func (s *errorStore) Add(member *domain.Member) error {
	return errors.New("mock store error")
}

func TestWebSocketHandler_HandleWebSocket_StoreError(t *testing.T) {
	// Create a test server with a handler that uses the error store
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := NewWebSocketHandler()
		handler.store = &errorStore{} // Replace store with error store
		handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Try to connect
	wsURL := "ws" + server.URL[4:]
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Error("Expected connection error")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
} 