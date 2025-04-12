package protocol

import (
	"encoding/json"
	"testing"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
)

// MockLogger implements logging.Logger for testing
type MockLogger struct{}

func (l *MockLogger) Debug(msg string, keyvals ...interface{})   {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})    {}
func (l *MockLogger) Error(msg string, keyvals ...interface{})   {}
func (l *MockLogger) With(keyvals ...interface{}) logging.Logger { return l }

func TestJoinRoom(t *testing.T) {
	sm := NewSignalingManager(&MockLogger{})

	// Create a join message
	msg := Message{
		Type:   Join,
		Room:   "test-room",
		Sender: "client-1",
	}

	// Encode the message
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// Process the message
	err = sm.ProcessMessage(msgJSON, "client-1", func(string, []byte) error { return nil })
	if err != nil {
		t.Fatalf("Process message failed: %v", err)
	}

	// Verify the client is in the room
	peers := sm.GetPeersInRoom("test-room")
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer in room, got %d", len(peers))
	}

	if peers[0] != "client-1" {
		t.Errorf("Expected client-1 in room, got %s", peers[0])
	}

	// Verify room exists
	if !sm.RoomExists("test-room") {
		t.Error("Expected room to exist")
	}
}

func TestLeaveRoom(t *testing.T) {
	sm := NewSignalingManager(&MockLogger{})

	// First join a room
	joinMsg := Message{
		Type:   Join,
		Room:   "test-room",
		Sender: "client-1",
	}
	joinJSON, _ := json.Marshal(joinMsg)
	sm.ProcessMessage(joinJSON, "client-1", func(string, []byte) error { return nil })

	// Create a leave message
	leaveMsg := Message{
		Type:   Leave,
		Room:   "test-room",
		Sender: "client-1",
	}
	leaveJSON, _ := json.Marshal(leaveMsg)

	// Process the leave message
	err := sm.ProcessMessage(leaveJSON, "client-1", func(string, []byte) error { return nil })
	if err != nil {
		t.Fatalf("Process leave message failed: %v", err)
	}

	// Verify the client is no longer in the room
	peers := sm.GetPeersInRoom("test-room")
	if len(peers) != 0 {
		t.Errorf("Expected 0 peers in room, got %d", len(peers))
	}

	// Verify room doesn't exist since it's empty
	if sm.RoomExists("test-room") {
		t.Error("Expected room to be removed after last client left")
	}
}

func TestRelayMessage(t *testing.T) {
	sm := NewSignalingManager(&MockLogger{})

	// First join two clients
	joinMsg1 := Message{
		Type:   Join,
		Room:   "test-room",
		Sender: "client-1",
	}
	joinJSON1, _ := json.Marshal(joinMsg1)
	sm.ProcessMessage(joinJSON1, "client-1", func(string, []byte) error { return nil })

	joinMsg2 := Message{
		Type:   Join,
		Room:   "test-room",
		Sender: "client-2",
	}
	joinJSON2, _ := json.Marshal(joinMsg2)
	sm.ProcessMessage(joinJSON2, "client-2", func(string, []byte) error { return nil })

	// Track message relay
	relayCount := 0
	relayClientID := ""
	relayMessageContent := []byte(nil)

	senderFunc := func(clientID string, message []byte) error {
		relayCount++
		relayClientID = clientID
		relayMessageContent = message
		return nil
	}

	// Create an offer message from client-1 to client-2
	offerPayload := json.RawMessage(`{"sdp":"test-sdp"}`)
	offerMsg := Message{
		Type:      Offer,
		Room:      "test-room",
		Sender:    "client-1",
		Recipient: "client-2",
		Payload:   offerPayload,
	}
	offerJSON, _ := json.Marshal(offerMsg)

	// Process the offer message
	err := sm.ProcessMessage(offerJSON, "client-1", senderFunc)
	if err != nil {
		t.Fatalf("Process offer message failed: %v", err)
	}

	// Verify message was relayed
	if relayCount != 1 {
		t.Errorf("Expected 1 message relay, got %d", relayCount)
	}

	if relayClientID != "client-2" {
		t.Errorf("Expected message relay to client-2, got %s", relayClientID)
	}

	// Decode the relayed message
	var relayedMsg Message
	err = json.Unmarshal(relayMessageContent, &relayedMsg)
	if err != nil {
		t.Fatalf("Failed to unmarshal relayed message: %v", err)
	}

	// Verify message content
	if relayedMsg.Type != Offer {
		t.Errorf("Expected message type %s, got %s", Offer, relayedMsg.Type)
	}

	if relayedMsg.Sender != "client-1" {
		t.Errorf("Expected sender client-1, got %s", relayedMsg.Sender)
	}

	if relayedMsg.Recipient != "client-2" {
		t.Errorf("Expected recipient client-2, got %s", relayedMsg.Recipient)
	}

	// Verify payload
	var payload struct {
		SDP string `json:"sdp"`
	}
	err = json.Unmarshal(relayedMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if payload.SDP != "test-sdp" {
		t.Errorf("Expected SDP test-sdp, got %s", payload.SDP)
	}
}

func TestRoomManagement(t *testing.T) {
	sm := NewSignalingManager(&MockLogger{})

	// Initially no rooms
	if sm.GetRoomCount() != 0 {
		t.Errorf("Expected 0 rooms, got %d", sm.GetRoomCount())
	}

	// Join multiple clients to multiple rooms
	clients := []string{"client-1", "client-2", "client-3", "client-4"}
	rooms := []string{"room-1", "room-1", "room-2", "room-2"}

	for i, client := range clients {
		joinMsg := Message{
			Type:   Join,
			Room:   rooms[i],
			Sender: client,
		}
		joinJSON, _ := json.Marshal(joinMsg)
		sm.ProcessMessage(joinJSON, client, func(string, []byte) error { return nil })
	}

	// Verify room count
	if sm.GetRoomCount() != 2 {
		t.Errorf("Expected 2 rooms, got %d", sm.GetRoomCount())
	}

	// Verify room populations
	room1Peers := sm.GetPeersInRoom("room-1")
	if len(room1Peers) != 2 {
		t.Errorf("Expected 2 peers in room-1, got %d", len(room1Peers))
	}

	room2Peers := sm.GetPeersInRoom("room-2")
	if len(room2Peers) != 2 {
		t.Errorf("Expected 2 peers in room-2, got %d", len(room2Peers))
	}

	// Make all clients leave room 1
	for i := 0; i < 2; i++ {
		leaveMsg := Message{
			Type:   Leave,
			Room:   "room-1",
			Sender: clients[i],
		}
		leaveJSON, _ := json.Marshal(leaveMsg)
		sm.ProcessMessage(leaveJSON, clients[i], func(string, []byte) error { return nil })
	}

	// Verify room 1 is gone
	if sm.RoomExists("room-1") {
		t.Error("Expected room-1 to be removed")
	}

	// Verify room count
	if sm.GetRoomCount() != 1 {
		t.Errorf("Expected 1 room, got %d", sm.GetRoomCount())
	}
}
