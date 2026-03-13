package service

import (
	"sync"

	"github.com/xyun1996/social_backend/services/identity/internal/domain"
)

// AccountStore persists account-to-player ownership for identity.
type AccountStore interface {
	UpsertAccount(accountID string, playerID string) error
}

// SessionStore persists issued identity sessions for introspection and refresh.
type SessionStore interface {
	SaveSession(session domain.Session) error
	GetSessionByRefreshToken(refreshToken string) (domain.Session, bool, error)
	GetSessionByAccessToken(accessToken string) (domain.Session, bool, error)
	DeleteSessionByRefreshToken(refreshToken string) error
}

type memoryAccountStore struct {
	accounts map[string]string
}

func newMemoryAccountStore() *memoryAccountStore {
	return &memoryAccountStore{accounts: make(map[string]string)}
}

func (s *memoryAccountStore) UpsertAccount(accountID string, playerID string) error {
	s.accounts[accountID] = playerID
	return nil
}

type memorySessionStore struct {
	byRefreshToken map[string]domain.Session
	byAccessToken  map[string]domain.Session
}

func newMemorySessionStore() *memorySessionStore {
	return &memorySessionStore{
		byRefreshToken: make(map[string]domain.Session),
		byAccessToken:  make(map[string]domain.Session),
	}
}

func (s *memorySessionStore) SaveSession(session domain.Session) error {
	s.byRefreshToken[session.RefreshToken] = session
	s.byAccessToken[session.AccessToken] = session
	return nil
}

func (s *memorySessionStore) GetSessionByRefreshToken(refreshToken string) (domain.Session, bool, error) {
	session, ok := s.byRefreshToken[refreshToken]
	return session, ok, nil
}

func (s *memorySessionStore) GetSessionByAccessToken(accessToken string) (domain.Session, bool, error) {
	session, ok := s.byAccessToken[accessToken]
	return session, ok, nil
}

func (s *memorySessionStore) DeleteSessionByRefreshToken(refreshToken string) error {
	session, ok := s.byRefreshToken[refreshToken]
	if ok {
		delete(s.byRefreshToken, refreshToken)
		delete(s.byAccessToken, session.AccessToken)
	}
	return nil
}

// memoryStores groups the default in-memory store implementations under one lock.
type memoryStores struct {
	mu       sync.RWMutex
	accounts *memoryAccountStore
	sessions *memorySessionStore
}

func newMemoryStores() *memoryStores {
	return &memoryStores{
		accounts: newMemoryAccountStore(),
		sessions: newMemorySessionStore(),
	}
}
