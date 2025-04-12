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

// DMCommand handles sending direct messages to a specific member
type DMCommand struct {
	Member    *domain.Member
	Recipient string
	Message   string
	Store     domain.MemberStore
}

func (c *DMCommand) Execute() error {
	// Find recipient member by name
	members := c.Store.List()
	var recipientMember *domain.Member
	
	for _, m := range members {
		if m.Name == c.Recipient {
			recipientMember = m
			break
		}
	}
	
	if recipientMember == nil {
		// Send error back to sender
		errorEvent := &wire.EventMessage{
			Event:   "error",
			Message: fmt.Sprintf("Member '%s' not found", c.Recipient),
		}
		return c.Member.Conn.WriteJSON(errorEvent)
	}
	
	// Create DM event
	dmEvent := wire.NewDMEventMessage(c.Member.Name, c.Message)
	
	// Send to recipient
	if err := recipientMember.Conn.WriteJSON(dmEvent); err != nil {
		log.Printf("Error sending DM to member %s: %v", recipientMember.ID, err)
		return err
	}
	
	// Send confirmation to sender
	confirmEvent := &wire.EventMessage{
		Event:   "dm_sent",
		Member:  c.Recipient,
		Message: c.Message,
	}
	if err := c.Member.Conn.WriteJSON(confirmEvent); err != nil {
		log.Printf("Error sending confirmation to member %s: %v", c.Member.ID, err)
	}
	
	log.Printf("📤 DM from %s to %s: %s", c.Member.Name, c.Recipient, c.Message)
	return nil
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
	case "dm":
		return &DMCommand{
			Member:    member,
			Recipient: msg.Recipient,
			Message:   msg.Message,
			Store:     store,
		}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", msg.Command)
	}
}
