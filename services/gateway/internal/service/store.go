package service

import (
	"slices"
	"sync"
)

// RealtimeSessionStore persists gateway-owned session runtime views.
type RealtimeSessionStore interface {
	SaveSession(session RealtimeSession) error
	GetSession(sessionID string) (RealtimeSession, bool, error)
	ListSessions() ([]RealtimeSession, error)
}

// SessionEventStore persists gateway-owned buffered session events.
type SessionEventStore interface {
	SaveEvents(sessionID string, events []ChatMessageEnvelope) error
	GetEvents(sessionID string) ([]ChatMessageEnvelope, error)
}

type memoryRealtimeSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]RealtimeSession
}

func newMemoryRealtimeSessionStore() *memoryRealtimeSessionStore {
	return &memoryRealtimeSessionStore{
		sessions: make(map[string]RealtimeSession),
	}
}

func (s *memoryRealtimeSessionStore) SaveSession(session RealtimeSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.SessionID] = session
	return nil
}

func (s *memoryRealtimeSessionStore) GetSession(sessionID string) (RealtimeSession, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	return session, ok, nil
}

func (s *memoryRealtimeSessionStore) ListSessions() ([]RealtimeSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]RealtimeSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	slices.SortFunc(sessions, func(a RealtimeSession, b RealtimeSession) int {
		switch {
		case a.SessionID < b.SessionID:
			return -1
		case a.SessionID > b.SessionID:
			return 1
		default:
			return 0
		}
	})
	return sessions, nil
}

type memorySessionEventStore struct {
	mu     sync.RWMutex
	events map[string][]ChatMessageEnvelope
}

func newMemorySessionEventStore() *memorySessionEventStore {
	return &memorySessionEventStore{
		events: make(map[string][]ChatMessageEnvelope),
	}
}

func (s *memorySessionEventStore) SaveEvents(sessionID string, events []ChatMessageEnvelope) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[sessionID] = append([]ChatMessageEnvelope(nil), events...)
	return nil
}

func (s *memorySessionEventStore) GetEvents(sessionID string) ([]ChatMessageEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ChatMessageEnvelope(nil), s.events[sessionID]...), nil
}
