package transport

import (
	"chat-server-go/domain"
	"errors"
	"sync"
)

// MemberStore manages the collection of connected members
type MemberStore struct {
	members map[string]*domain.Member
	mu      sync.RWMutex
}

// NewMemberStore creates a new MemberStore instance
func NewMemberStore() *MemberStore {
	return &MemberStore{
		members: make(map[string]*domain.Member),
	}
}

// Add adds a new member to the store
func (s *MemberStore) Add(member *domain.Member) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[member.ID] = member
	return nil
}

// Remove removes a member from the store by ID
func (s *MemberStore) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.members, id)
	return nil
}

// Get retrieves a member from the store by ID
func (s *MemberStore) Get(id string) (*domain.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.members[id]
	if !ok {
		return nil, errors.New("member not found")
	}
	return member, nil
}

// List returns a slice of all members in the store
func (s *MemberStore) List() []*domain.Member {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members := make([]*domain.Member, 0, len(s.members))
	for _, m := range s.members {
		members = append(members, m)
	}
	return members
} 