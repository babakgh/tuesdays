package transport

import "github.com/gorilla/websocket"

// Member represents a connected chat member
type Member struct {
	ID   string
	Name string
	Conn *websocket.Conn
} 