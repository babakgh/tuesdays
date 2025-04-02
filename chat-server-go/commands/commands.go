package commands

import (
	"fmt"
	"log"

	"chat-server-go/domain"
	"chat-server-go/wire"
)

// BroadcastCommand handles broadcasting messages to all members
type BroadcastCommand struct {
	Member  *domain.Member
	Message string
	Store   domain.MemberStore
}

func (c *BroadcastCommand) Execute() error {
	event := &wire.EventMessage{
		Event:   "broadcast",
		Member:  c.Member.Name,
		Message: c.Message,
	}

	// Get all members and broadcast to each
	members := c.Store.List()
	for _, member := range members {
		if err := member.Conn.WriteJSON(event); err != nil {
			log.Printf("Error broadcasting to member %s: %v", member.ID, err)
			continue
		}
	}

	log.Printf("📤 Broadcast from %s: %s", c.Member.Name, c.Message)
	return nil
}

// ListCommand handles listing all connected members
type ListCommand struct {
	Member *domain.Member
	Store  domain.MemberStore
}

func (c *ListCommand) Execute() error {
	members := c.Store.List()
	memberNames := make([]string, len(members))
	for i, m := range members {
		memberNames[i] = m.Name
	}

	event := &wire.EventMessage{
		Event:   "list",
		Members: memberNames,
	}
	return c.Member.Conn.WriteJSON(event)
}

// MeCommand handles returning the current member's information
type MeCommand struct {
	Member *domain.Member
}

func (c *MeCommand) Execute() error {
	event := &wire.EventMessage{
		Event:  "me",
		Member: c.Member.Name,
		Data:   map[string]string{"id": c.Member.ID},
	}
	return c.Member.Conn.WriteJSON(event)
}

// CommandFactory creates the appropriate command based on the message type
func CommandFactory(msg *wire.CommandMessage, member *domain.Member, store domain.MemberStore) (domain.Command, error) {
	switch msg.Command {
	case "broadcast":
		return &BroadcastCommand{
			Member:  member,
			Message: msg.Message,
			Store:   store,
		}, nil
	case "list":
		return &ListCommand{
			Member: member,
			Store:  store,
		}, nil
	case "me":
		return &MeCommand{
			Member: member,
		}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", msg.Command)
	}
}
