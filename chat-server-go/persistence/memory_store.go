package persistence

import (
	"errors"
	"sync"

	"chat-server-go/domain"
)

// MemoryStore implements the MemberStore interface using an in-memory map
type MemoryStore struct {
	mu      sync.RWMutex
	members map[string]*domain.Member
}

// NewMemoryStore creates a new instance of MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		members: make(map[string]*domain.Member),
	}
}

// Add adds a new member to the store
func (s *MemoryStore) Add(member *domain.Member) error {
	if member == nil {
		return errors.New("member cannot be nil")
	}
	if member.ID == "" {
		return errors.New("member ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.members[member.ID]; exists {
		return errors.New("member already exists")
	}

	s.members[member.ID] = member
	return nil
}

// Remove removes a member from the store
func (s *MemoryStore) Remove(memberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.members[memberID]; !exists {
		return errors.New("member not found")
	}

	delete(s.members, memberID)
	return nil
}

// Get retrieves a member by ID
func (s *MemoryStore) Get(memberID string) (*domain.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	member, exists := s.members[memberID]
	if !exists {
		return nil, errors.New("member not found")
	}

	return member, nil
}

// List returns all connected members
func (s *MemoryStore) List() []*domain.Member {
	s.mu.RLock()
	defer s.mu.RUnlock()

	members := make([]*domain.Member, 0, len(s.members))
	for _, member := range s.members {
		members = append(members, member)
	}
	return members
}
