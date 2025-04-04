package persistence

import (
	"fmt"
	"sync"
	"testing"

	"chat-server-go/domain"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	t.Run("Add member", func(t *testing.T) {
		member := &domain.Member{
			ID:   "test1",
			Name: "Test User 1",
		}

		// Test adding a new member
		err := store.Add(member)
		assert.NoError(t, err)

		// Test adding a duplicate member
		err = store.Add(member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member already exists")
	})

	t.Run("Get member", func(t *testing.T) {
		// Test getting an existing member
		member, err := store.Get("test1")
		assert.NoError(t, err)
		assert.Equal(t, "test1", member.ID)
		assert.Equal(t, "Test User 1", member.Name)

		// Test getting a non-existent member
		member, err = store.Get("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member not found")
	})

	t.Run("List members", func(t *testing.T) {
		// Add another member
		member2 := &domain.Member{
			ID:   "test2",
			Name: "Test User 2",
		}
		err := store.Add(member2)
		assert.NoError(t, err)

		// Test listing all members
		members := store.List()
		assert.Len(t, members, 2)

		// Verify members are in the list
		found := make(map[string]bool)
		for _, m := range members {
			found[m.ID] = true
		}
		assert.True(t, found["test1"])
		assert.True(t, found["test2"])
	})

	t.Run("Remove member", func(t *testing.T) {
		// Test removing an existing member
		err := store.Remove("test1")
		assert.NoError(t, err)

		// Verify member is removed
		_, err = store.Get("test1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member not found")

		// Test removing a non-existent member
		err = store.Remove("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member not found")
	})
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()
	var wg sync.WaitGroup

	// Test concurrent reads
	t.Run("concurrent reads", func(t *testing.T) {
		member := &domain.Member{ID: "test1", Name: "Test User 1"}
		err := store.Add(member)
		assert.NoError(t, err)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := store.Get("test1")
				assert.NoError(t, err)
				_ = store.List()
			}()
		}
		wg.Wait()
	})

	// Test concurrent writes
	t.Run("concurrent writes", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				member := &domain.Member{
					ID:   fmt.Sprintf("test%d", id),
					Name: fmt.Sprintf("Test User %d", id),
				}
				err := store.Add(member)
				if err == nil {
					err = store.Remove(member.ID)
					assert.NoError(t, err)
				}
			}(i)
		}
		wg.Wait()
	})
}

func TestMemoryStore_EdgeCases(t *testing.T) {
	store := NewMemoryStore()

	t.Run("nil member", func(t *testing.T) {
		err := store.Add(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member cannot be nil")
	})

	t.Run("empty member ID", func(t *testing.T) {
		member := &domain.Member{ID: "", Name: "Test User"}
		err := store.Add(member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member ID cannot be empty")
	})

	t.Run("empty store list", func(t *testing.T) {
		members := store.List()
		assert.Empty(t, members)
	})
}
