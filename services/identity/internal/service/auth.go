package service

import (
	"fmt"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/identity/internal/domain"
)

const (
	accessTokenBytes  = 16
	refreshTokenBytes = 24
	accessTokenTTL    = time.Hour
)

// AuthService provides a local in-memory auth flow for early development.
type AuthService struct {
	mu              sync.RWMutex
	refreshSessions map[string]domain.Session
	accessSessions  map[string]domain.Session
	now             func() time.Time
}

// NewAuthService constructs an in-memory auth service.
func NewAuthService() *AuthService {
	return &AuthService{
		refreshSessions: make(map[string]domain.Session),
		accessSessions:  make(map[string]domain.Session),
		now:             time.Now,
	}
}

// Login issues a token pair for the given account and player identifiers.
func (s *AuthService) Login(accountID string, playerID string) (domain.TokenPair, *apperrors.Error) {
	if accountID == "" || playerID == "" {
		err := apperrors.New("invalid_request", "account_id and player_id are required", 400)
		return domain.TokenPair{}, &err
	}

	pair, session, err := s.issueTokens(accountID, playerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}

	s.mu.Lock()
	s.refreshSessions[session.RefreshToken] = session
	s.accessSessions[session.AccessToken] = session
	s.mu.Unlock()

	return pair, nil
}

// Refresh rotates the token pair for an existing refresh token.
func (s *AuthService) Refresh(refreshToken string) (domain.TokenPair, *apperrors.Error) {
	if refreshToken == "" {
		err := apperrors.New("invalid_request", "refresh_token is required", 400)
		return domain.TokenPair{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.refreshSessions[refreshToken]
	if !ok {
		err := apperrors.New("unauthorized", "refresh token is invalid", 401)
		return domain.TokenPair{}, &err
	}

	delete(s.refreshSessions, refreshToken)
	delete(s.accessSessions, session.AccessToken)

	pair, newSession, err := s.issueTokens(session.AccountID, session.PlayerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}

	s.refreshSessions[newSession.RefreshToken] = newSession
	s.accessSessions[newSession.AccessToken] = newSession
	return pair, nil
}

// Introspect resolves an access token into the authenticated subject.
func (s *AuthService) Introspect(accessToken string) (domain.Subject, *apperrors.Error) {
	if accessToken == "" {
		err := apperrors.New("invalid_request", "access token is required", 400)
		return domain.Subject{}, &err
	}

	s.mu.RLock()
	session, ok := s.accessSessions[accessToken]
	s.mu.RUnlock()
	if !ok {
		err := apperrors.New("unauthorized", "access token is invalid", 401)
		return domain.Subject{}, &err
	}

	return domain.Subject{
		AccountID: session.AccountID,
		PlayerID:  session.PlayerID,
	}, nil
}

func (s *AuthService) issueTokens(accountID string, playerID string) (domain.TokenPair, domain.Session, error) {
	accessToken, err := idgen.Token(accessTokenBytes)
	if err != nil {
		return domain.TokenPair{}, domain.Session{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := idgen.Token(refreshTokenBytes)
	if err != nil {
		return domain.TokenPair{}, domain.Session{}, fmt.Errorf("generate refresh token: %w", err)
	}

	session := domain.Session{
		AccessToken:  accessToken,
		AccountID:    accountID,
		PlayerID:     playerID,
		RefreshToken: refreshToken,
		ExpiresAt:    s.now().Add(accessTokenTTL),
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    time.Duration(accessTokenTTL.Seconds()),
		AccountID:    accountID,
		PlayerID:     playerID,
	}, session, nil
}
