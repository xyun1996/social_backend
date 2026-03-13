package service

import (
	"sync"

	"github.com/xyun1996/social_backend/services/invite/internal/domain"
)

// InviteStore persists invite lifecycle state.
type InviteStore interface {
	ListInvites() ([]domain.Invite, error)
	SaveInvite(invite domain.Invite) error
	GetInvite(inviteID string) (domain.Invite, bool, error)
}

type memoryInviteStore struct {
	mu      sync.RWMutex
	invites map[string]domain.Invite
}

func newMemoryInviteStore() *memoryInviteStore {
	return &memoryInviteStore{
		invites: make(map[string]domain.Invite),
	}
}

func (s *memoryInviteStore) ListInvites() ([]domain.Invite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	invites := make([]domain.Invite, 0, len(s.invites))
	for _, invite := range s.invites {
		invites = append(invites, invite)
	}
	return invites, nil
}

func (s *memoryInviteStore) SaveInvite(invite domain.Invite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invites[invite.ID] = invite
	return nil
}

func (s *memoryInviteStore) GetInvite(inviteID string) (domain.Invite, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	invite, ok := s.invites[inviteID]
	return invite, ok, nil
}
