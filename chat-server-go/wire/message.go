package wire

import (
	"encoding/json"
	"fmt"
)

// CommandMessage represents the structure of incoming command messages
type CommandMessage struct {
	Command string          `json:"command"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// EventMessage represents the structure of outgoing event messages
type EventMessage struct {
	Event   string      `json:"event"`
	Member  string      `json:"member,omitempty"`
	Message string      `json:"message,omitempty"`
	Members []string    `json:"members,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ParseCommand parses a JSON message into a CommandMessage
func ParseCommand(data []byte) (*CommandMessage, error) {
	var cmd CommandMessage
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}
	return &cmd, nil
}

// NewEventMessage creates a new EventMessage with the given parameters
func NewEventMessage(event string, member string, message string) *EventMessage {
	return &EventMessage{
		Event:   event,
		Member:  member,
		Message: message,
	}
}

// NewListEventMessage creates a new EventMessage for the list command
func NewListEventMessage(members []string) *EventMessage {
	return &EventMessage{
		Event:   "list",
		Members: members,
	}
}

// NewMeEventMessage creates a new EventMessage for the me command
func NewMeEventMessage(member string, id string) *EventMessage {
	return &EventMessage{
		Event:  "me",
		Member: member,
		Data: map[string]string{
			"id": id,
		},
	}
}
