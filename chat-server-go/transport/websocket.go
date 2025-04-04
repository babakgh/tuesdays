package transport

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"chat-server-go/commands"
	"chat-server-go/domain"
	"chat-server-go/persistence"
	"chat-server-go/wire"

	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections and message routing
type WebSocketHandler struct {
	store    domain.MemberStore
	memberID uint64 // Atomic counter for generating unique member IDs
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		store:    persistence.NewMemoryStore(),
		memberID: 0,
	}
}

// HandleWebSocket handles the WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Generate unique member ID and name
	memberID := atomic.AddUint64(&h.memberID, 1)
	memberName := fmt.Sprintf("member%d", memberID)

	// Create a temporary member to test store availability
	tempMember := &domain.Member{
		ID:   memberName,
		Name: memberName,
	}

	// Test if we can add the member
	if err := h.store.Add(tempMember); err != nil {
		log.Printf("Failed to add member: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.store.Remove(tempMember.ID) // Remove the temporary member

	// Now upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	member := &domain.Member{
		ID:   memberName,
		Name: memberName,
		Conn: conn,
	}

	// Add member to store
	if err := h.store.Add(member); err != nil {
		log.Printf("Failed to add member: %v", err)
		conn.Close()
		return
	}

	log.Printf("🔌 Member %s connected", memberName)

	// Send welcome messages
	h.sendWelcomeMessages(member)

	// Start message handling loop
	go h.handleMessages(member)
}

func (h *WebSocketHandler) sendWelcomeMessages(member *domain.Member) {
	// Send me command
	meCmd := &commands.MeCommand{Member: member}
	meCmd.Execute()

	// Send join broadcast
	joinEvent := wire.NewEventMessage("broadcast", "", fmt.Sprintf("%s has joined!", member.Name))
	members := h.store.List()
	for _, m := range members {
		m.Conn.WriteJSON(joinEvent)
	}
}

func (h *WebSocketHandler) handleMessages(member *domain.Member) {
	defer func() {
		h.store.Remove(member.ID)
		member.Conn.Close()
		log.Printf("🔌 Member %s disconnected", member.Name)
	}()

	for {
		_, message, err := member.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Parse command message
		cmdMsg, err := wire.ParseCommand(message)
		if err != nil {
			log.Printf("Error parsing command: %v", err)
			continue
		}

		// Create and execute command
		cmd, err := commands.CommandFactory(cmdMsg, member, h.store)
		if err != nil {
			log.Printf("Error creating command: %v", err)
			continue
		}

		if err := cmd.Execute(); err != nil {
			log.Printf("Error executing command: %v", err)
		}
	}
}
