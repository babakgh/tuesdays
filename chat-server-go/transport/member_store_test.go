package transport

import (
	"testing"

	"chat-server-go/domain"
)

func TestMemberStore(t *testing.T) {
	store := NewMemberStore()

	// Test Add
	mockConn := newMockWebSocketConn()
	member := &domain.Member{
		ID:   "1",
		Name: "test",
		Conn: mockConn,
	}
	store.Add(member)

	// Test List
	members := store.List()
	if len(members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(members))
	}
	if members[0].ID != "1" {
		t.Errorf("Expected member ID 1, got %s", members[0].ID)
	}

	// Test Remove
	store.Remove("1")
	members = store.List()
	if len(members) != 0 {
		t.Errorf("Expected 0 members after removal, got %d", len(members))
	}
} 