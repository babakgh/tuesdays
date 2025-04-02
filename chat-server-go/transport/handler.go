package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Member struct {
	ID   string
	Name string
	Conn *websocket.Conn
}

type MemberStore struct {
	members map[string]*Member
}

func NewMemberStore() *MemberStore {
	return &MemberStore{
		members: make(map[string]*Member),
	}
}

func (s *MemberStore) Add(member *Member) {
	s.members[member.ID] = member
}

func (s *MemberStore) Remove(id string) {
	delete(s.members, id)
}

func (s *MemberStore) List() []*Member {
	list := make([]*Member, 0, len(s.members))
	for _, m := range s.members {
		list = append(list, m)
	}
	return list
}

type Handler struct {
	store    *MemberStore
	memberID uint64
}

func NewHandler() *Handler {
	return &Handler{
		store: NewMemberStore(),
	}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	memberID := atomic.AddUint64(&h.memberID, 1)
	memberName := fmt.Sprintf("member%d", memberID)

	member := &Member{
		ID:   memberName,
		Name: memberName,
		Conn: conn,
	}

	h.store.Add(member)
	log.Printf("🔌 Member %s connected", memberName)

	// Send welcome messages
	h.sendMeEvent(member)
	h.broadcastJoin(member)

	// Handle messages
	go h.handleMessages(member)
}

func (h *Handler) sendMeEvent(member *Member) {
	event := map[string]interface{}{
		"event":  "me",
		"member": member.Name,
		"id":     member.ID,
	}
	member.Conn.WriteJSON(event)
}

func (h *Handler) broadcastJoin(member *Member) {
	event := map[string]interface{}{
		"event":   "broadcast",
		"member":  "",
		"message": fmt.Sprintf("%s has joined!", member.Name),
	}
	h.broadcast(event)
}

func (h *Handler) broadcast(event interface{}) {
	for _, member := range h.store.List() {
		member.Conn.WriteJSON(event)
	}
}

func (h *Handler) handleMessages(member *Member) {
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

		var cmd struct {
			Command string `json:"command"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Error parsing command: %v", err)
			continue
		}

		switch cmd.Command {
		case "broadcast":
			event := map[string]interface{}{
				"event":   "broadcast",
				"member":  member.Name,
				"message": cmd.Message,
			}
			h.broadcast(event)
			log.Printf("📤 Broadcast from %s: %s", member.Name, cmd.Message)

		case "list":
			members := h.store.List()
			names := make([]string, len(members))
			for i, m := range members {
				names[i] = m.Name
			}
			event := map[string]interface{}{
				"event":   "list",
				"members": names,
			}
			member.Conn.WriteJSON(event)

		case "me":
			h.sendMeEvent(member)
		}
	}
}
