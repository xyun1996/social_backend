package service

import (
	"sync"

	"github.com/xyun1996/social_backend/services/presence/internal/domain"
)

// PresenceStore persists the latest presence snapshot for a player.
type PresenceStore interface {
	SavePresence(presence domain.Presence) error
	GetPresence(playerID string) (domain.Presence, bool, error)
}

type memoryPresenceStore struct {
	mu        sync.RWMutex
	presences map[string]domain.Presence
}

func newMemoryPresenceStore() *memoryPresenceStore {
	return &memoryPresenceStore{
		presences: make(map[string]domain.Presence),
	}
}

func (s *memoryPresenceStore) SavePresence(presence domain.Presence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presences[presence.PlayerID] = presence
	return nil
}

func (s *memoryPresenceStore) GetPresence(playerID string) (domain.Presence, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	presence, ok := s.presences[playerID]
	return presence, ok, nil
}
