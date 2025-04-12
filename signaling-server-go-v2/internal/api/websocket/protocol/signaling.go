package protocol

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
)

// MessageType defines the type of WebRTC signaling message
type MessageType string

const (
	// Offer message - sent by a peer to initiate a connection
	Offer MessageType = "offer"

	// Answer message - sent in response to an offer
	Answer MessageType = "answer"

	// ICECandidate message - sent when a new ICE candidate is discovered
	ICECandidate MessageType = "ice-candidate"

	// Join message - sent when a peer wants to join a room
	Join MessageType = "join"

	// Leave message - sent when a peer wants to leave a room
	Leave MessageType = "leave"
)

// Message represents a signaling message
type Message struct {
	Type      MessageType     `json:"type"`
	Room      string          `json:"room,omitempty"`
	Sender    string          `json:"sender"`
	Recipient string          `json:"recipient,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// Room represents a signaling room with connected peers
type Room struct {
	ID    string
	Peers map[string]struct{}
	mutex sync.RWMutex
}

// SignalingManager handles signaling message routing and room management
type SignalingManager struct {
	rooms  map[string]*Room
	mutex  sync.RWMutex
	logger logging.Logger
}

// NewSignalingManager creates a new SignalingManager
func NewSignalingManager(logger logging.Logger) *SignalingManager {
	return &SignalingManager{
		rooms:  make(map[string]*Room),
		logger: logger.With("component", "signaling"),
	}
}

// ProcessMessage processes an incoming signaling message
func (sm *SignalingManager) ProcessMessage(message []byte, clientID string, sender func(string, []byte) error) error {
	// Parse the message
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		sm.logger.Error("Failed to unmarshal message", "error", err)
		return fmt.Errorf("invalid message format: %w", err)
	}

	// Set the sender ID
	msg.Sender = clientID

	// Handle the message based on its type
	switch msg.Type {
	case Join:
		return sm.handleJoin(msg, clientID)
	case Leave:
		return sm.handleLeave(msg, clientID)
	case Offer, Answer, ICECandidate:
		return sm.relayMessage(msg, sender)
	default:
		sm.logger.Warn("Unknown message type", "type", msg.Type)
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleJoin adds a client to a room
func (sm *SignalingManager) handleJoin(msg Message, clientID string) error {
	if msg.Room == "" {
		return fmt.Errorf("room ID is required for join messages")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Get or create the room
	room, ok := sm.rooms[msg.Room]
	if !ok {
		room = &Room{
			ID:    msg.Room,
			Peers: make(map[string]struct{}),
		}
		sm.rooms[msg.Room] = room
	}

	// Add the client to the room
	room.mutex.Lock()
	defer room.mutex.Unlock()

	room.Peers[clientID] = struct{}{}

	sm.logger.Info("Client joined room", "client_id", clientID, "room_id", msg.Room)
	return nil
}

// handleLeave removes a client from a room
func (sm *SignalingManager) handleLeave(msg Message, clientID string) error {
	if msg.Room == "" {
		return fmt.Errorf("room ID is required for leave messages")
	}

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Get the room
	room, ok := sm.rooms[msg.Room]
	if !ok {
		return fmt.Errorf("room not found: %s", msg.Room)
	}

	// Remove the client from the room
	room.mutex.Lock()
	delete(room.Peers, clientID)

	// If the room is empty, remove it
	if len(room.Peers) == 0 {
		sm.mutex.RUnlock()
		sm.mutex.Lock()
		delete(sm.rooms, msg.Room)
		sm.mutex.Unlock()
		sm.mutex.RLock()
	}
	room.mutex.Unlock()

	sm.logger.Info("Client left room", "client_id", clientID, "room_id", msg.Room)
	return nil
}

// relayMessage relays a message to its intended recipient
func (sm *SignalingManager) relayMessage(msg Message, sender func(string, []byte) error) error {
	if msg.Recipient == "" {
		return fmt.Errorf("recipient is required for relay messages")
	}

	// Marshal the message
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		sm.logger.Error("Failed to marshal message", "error", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send the message to the recipient
	if err := sender(msg.Recipient, messageJSON); err != nil {
		sm.logger.Error("Failed to send message", "error", err, "recipient", msg.Recipient)
		return fmt.Errorf("failed to send message: %w", err)
	}

	sm.logger.Debug("Message relayed", "from", msg.Sender, "to", msg.Recipient, "type", msg.Type)
	return nil
}

// GetPeersInRoom returns all peers in a room
func (sm *SignalingManager) GetPeersInRoom(roomID string) []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	room, ok := sm.rooms[roomID]
	if !ok {
		return []string{}
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	peers := make([]string, 0, len(room.Peers))
	for peer := range room.Peers {
		peers = append(peers, peer)
	}

	return peers
}

// RoomExists checks if a room exists
func (sm *SignalingManager) RoomExists(roomID string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	_, ok := sm.rooms[roomID]
	return ok
}

// GetRoomCount returns the number of active rooms
func (sm *SignalingManager) GetRoomCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return len(sm.rooms)
}
