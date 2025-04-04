package wire

import (
	"encoding/json"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *CommandMessage
		wantErr bool
	}{
		{
			name:  "valid command with message",
			input: `{"command": "join", "message": "hello"}`,
			want: &CommandMessage{
				Command: "join",
				Message: "hello",
			},
			wantErr: false,
		},
		{
			name:  "valid command with data",
			input: `{"command": "me", "data": {"id": "123"}}`,
			want: &CommandMessage{
				Command: "me",
				Data:    json.RawMessage(`{"id": "123"}`),
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{"command": "join", "message": "hello"`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Command != tt.want.Command {
					t.Errorf("ParseCommand() Command = %v, want %v", got.Command, tt.want.Command)
				}
				if got.Message != tt.want.Message {
					t.Errorf("ParseCommand() Message = %v, want %v", got.Message, tt.want.Message)
				}
				if string(got.Data) != string(tt.want.Data) {
					t.Errorf("ParseCommand() Data = %v, want %v", string(got.Data), string(tt.want.Data))
				}
			}
		})
	}
}

func TestNewEventMessage(t *testing.T) {
	msg := NewEventMessage("join", "user1", "hello")
	if msg.Event != "join" {
		t.Errorf("NewEventMessage() Event = %v, want %v", msg.Event, "join")
	}
	if msg.Member != "user1" {
		t.Errorf("NewEventMessage() Member = %v, want %v", msg.Member, "user1")
	}
	if msg.Message != "hello" {
		t.Errorf("NewEventMessage() Message = %v, want %v", msg.Message, "hello")
	}
}

func TestNewListEventMessage(t *testing.T) {
	members := []string{"user1", "user2", "user3"}
	msg := NewListEventMessage(members)
	if msg.Event != "list" {
		t.Errorf("NewListEventMessage() Event = %v, want %v", msg.Event, "list")
	}
	if len(msg.Members) != len(members) {
		t.Errorf("NewListEventMessage() Members length = %v, want %v", len(msg.Members), len(members))
	}
	for i, member := range members {
		if msg.Members[i] != member {
			t.Errorf("NewListEventMessage() Members[%d] = %v, want %v", i, msg.Members[i], member)
		}
	}
}

func TestNewMeEventMessage(t *testing.T) {
	msg := NewMeEventMessage("user1", "123")
	if msg.Event != "me" {
		t.Errorf("NewMeEventMessage() Event = %v, want %v", msg.Event, "me")
	}
	if msg.Member != "user1" {
		t.Errorf("NewMeEventMessage() Member = %v, want %v", msg.Member, "user1")
	}
	data, ok := msg.Data.(map[string]string)
	if !ok {
		t.Errorf("NewMeEventMessage() Data type = %T, want map[string]string", msg.Data)
	}
	if data["id"] != "123" {
		t.Errorf("NewMeEventMessage() Data[id] = %v, want %v", data["id"], "123")
	}
} 