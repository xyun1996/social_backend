package service

import (
	"slices"
	"sync"

	"github.com/xyun1996/social_backend/services/party/internal/domain"
)

// PartyStore persists party aggregate state.
type PartyStore interface {
	SaveParty(party domain.Party) error
	GetParty(partyID string) (domain.Party, bool, error)
	ListParties() ([]domain.Party, error)
}

// ReadyStateStore persists per-party ready state snapshots.
type ReadyStateStore interface {
	SaveReadyState(state domain.ReadyState) error
	ListReadyStates(partyID string) ([]domain.ReadyState, error)
	DeleteReadyState(partyID string, playerID string) error
}

// QueueStateStore persists active social queue enrollment per party.
type QueueStateStore interface {
	SaveQueueState(state domain.QueueState) error
	GetQueueState(partyID string) (domain.QueueState, bool, error)
	DeleteQueueState(partyID string) error
	SaveQueueAssignment(assignment domain.QueueAssignment) error
	GetQueueAssignment(partyID string) (domain.QueueAssignment, bool, error)
	DeleteQueueAssignment(partyID string) error
}

type memoryPartyStore struct {
	mu      sync.RWMutex
	parties map[string]domain.Party
}

func newMemoryPartyStore() *memoryPartyStore {
	return &memoryPartyStore{
		parties: make(map[string]domain.Party),
	}
}

func (s *memoryPartyStore) SaveParty(party domain.Party) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.parties[party.ID] = party
	return nil
}

func (s *memoryPartyStore) GetParty(partyID string) (domain.Party, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	party, ok := s.parties[partyID]
	return party, ok, nil
}

func (s *memoryPartyStore) ListParties() ([]domain.Party, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parties := make([]domain.Party, 0, len(s.parties))
	for _, party := range s.parties {
		parties = append(parties, party)
	}
	slices.SortFunc(parties, func(a domain.Party, b domain.Party) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})
	return parties, nil
}

type memoryReadyStateStore struct {
	mu     sync.RWMutex
	states map[string]map[string]domain.ReadyState
}

type memoryQueueStateStore struct {
	mu          sync.RWMutex
	queues      map[string]domain.QueueState
	assignments map[string]domain.QueueAssignment
}

func newMemoryReadyStateStore() *memoryReadyStateStore {
	return &memoryReadyStateStore{
		states: make(map[string]map[string]domain.ReadyState),
	}
}

func newMemoryQueueStateStore() *memoryQueueStateStore {
	return &memoryQueueStateStore{
		queues:      make(map[string]domain.QueueState),
		assignments: make(map[string]domain.QueueAssignment),
	}
}

func (s *memoryReadyStateStore) SaveReadyState(state domain.ReadyState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.states[state.PartyID] == nil {
		s.states[state.PartyID] = make(map[string]domain.ReadyState)
	}
	s.states[state.PartyID][state.PlayerID] = state
	return nil
}

func (s *memoryReadyStateStore) ListReadyStates(partyID string) ([]domain.ReadyState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	partyStates := s.states[partyID]
	states := make([]domain.ReadyState, 0, len(partyStates))
	for _, state := range partyStates {
		states = append(states, state)
	}
	slices.SortFunc(states, func(a domain.ReadyState, b domain.ReadyState) int {
		switch {
		case a.PlayerID < b.PlayerID:
			return -1
		case a.PlayerID > b.PlayerID:
			return 1
		default:
			return 0
		}
	})
	return states, nil
}

func (s *memoryReadyStateStore) DeleteReadyState(partyID string, playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.states[partyID] == nil {
		return nil
	}

	delete(s.states[partyID], playerID)
	if len(s.states[partyID]) == 0 {
		delete(s.states, partyID)
	}
	return nil
}

func (s *memoryQueueStateStore) SaveQueueState(state domain.QueueState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queues[state.PartyID] = state
	return nil
}

func (s *memoryQueueStateStore) GetQueueState(partyID string) (domain.QueueState, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.queues[partyID]
	return state, ok, nil
}

func (s *memoryQueueStateStore) DeleteQueueState(partyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.queues, partyID)
	return nil
}

func (s *memoryQueueStateStore) SaveQueueAssignment(assignment domain.QueueAssignment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assignments[assignment.PartyID] = assignment
	return nil
}

func (s *memoryQueueStateStore) GetQueueAssignment(partyID string) (domain.QueueAssignment, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	assignment, ok := s.assignments[partyID]
	return assignment, ok, nil
}

func (s *memoryQueueStateStore) DeleteQueueAssignment(partyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.assignments, partyID)
	return nil
}
