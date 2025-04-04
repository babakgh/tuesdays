package domain

// Command defines the interface all client commands must implement
// to execute logic against the server context.
type Command interface {
	Execute() error
}

// Event defines the interface all server-pushed messages should implement
type Event interface {
	Name() string         // e.g., "broadcast", "system"
	Payload() interface{} // the data to be encoded into the JSON response
}

// WebSocketConn is an interface that abstracts the websocket.Conn methods we need
type WebSocketConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteJSON(v interface{}) error
	Close() error
}

// Member represents a connected chat member
type Member struct {
	ID   string
	Name string
	Conn WebSocketConn
}

// MemberStore defines the interface for managing connected members
type MemberStore interface {
	Add(member *Member) error
	Remove(memberID string) error
	Get(memberID string) (*Member, error)
	List() []*Member
}

// Broadcaster defines the interface for broadcasting messages to members
type Broadcaster interface {
	Broadcast(event Event) error
	BroadcastTo(memberID string, event Event) error
}

// CommandHandler defines the interface for handling different types of commands
type CommandHandler interface {
	Handle(command Command) error
}
