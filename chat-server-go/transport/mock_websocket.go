package transport

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
)

// WebSocketConn is an interface that abstracts the websocket.Conn methods we need for testing
type WebSocketConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteJSON(v interface{}) error
	Close() error
}

// mockWebSocketConn is a mock implementation of WebSocketConn for testing
type mockWebSocketConn struct {
	readChan  chan []byte
	writeChan chan []byte
	closeChan chan struct{}
	closed    bool
}

func newMockWebSocketConn() WebSocketConn {
	return &mockWebSocketConn{
		readChan:  make(chan []byte, 10),  // Buffer size of 10 for test messages
		writeChan: make(chan []byte, 10),  // Buffer size of 10 for test messages
		closeChan: make(chan struct{}, 1),
		closed:    false,
	}
}

// ReadMessage implements the WebSocketConn ReadMessage method
func (m *mockWebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	select {
	case msg := <-m.readChan:
		return websocket.TextMessage, msg, nil
	case <-m.closeChan:
		return 0, nil, websocket.ErrCloseSent
	}
}

// WriteJSON implements the WebSocketConn WriteJSON method
func (m *mockWebSocketConn) WriteJSON(v interface{}) error {
	if m.closed {
		return websocket.ErrCloseSent
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// Use non-blocking send to avoid panic on closed channel
	select {
	case m.writeChan <- data:
		return nil
	case <-m.closeChan:
		return websocket.ErrCloseSent
	default:
		return errors.New("write channel full or closed")
	}
}

// Close implements the WebSocketConn Close method
func (m *mockWebSocketConn) Close() error {
	if !m.closed {
		m.closed = true
		close(m.closeChan)
	}
	return nil
} 