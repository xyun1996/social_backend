package service

import (
	"fmt"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/identity/internal/domain"
)

const (
	accessTokenBytes       = 16
	refreshTokenBytes      = 24
	defaultAccessTokenTTL  = time.Hour
	defaultRefreshTokenTTL = 7 * 24 * time.Hour
)

// Options configures token lifecycle defaults for the identity service.
type Options struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// AuthService provides a local in-memory auth flow for early development.
type AuthService struct {
	accounts   AccountStore
	sessions   SessionStore
	now        func() time.Time
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewAuthService constructs an in-memory auth service.
func NewAuthService() *AuthService {
	stores := newMemoryStores()
	return NewAuthServiceWithOptions(accountStoreWithLock(stores), sessionStoreWithLock(stores), Options{})
}

// NewAuthServiceWithStores constructs an auth service with custom persistence stores.
func NewAuthServiceWithStores(accounts AccountStore, sessions SessionStore) *AuthService {
	return NewAuthServiceWithOptions(accounts, sessions, Options{})
}

// NewAuthServiceWithOptions constructs an auth service with configurable stores and token TTLs.
func NewAuthServiceWithOptions(accounts AccountStore, sessions SessionStore, options Options) *AuthService {
	if accounts == nil || sessions == nil {
		return NewAuthService()
	}

	accessTTL := options.AccessTokenTTL
	if accessTTL <= 0 {
		accessTTL = defaultAccessTokenTTL
	}

	refreshTTL := options.RefreshTokenTTL
	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTokenTTL
	}

	return &AuthService{
		accounts:   accounts,
		sessions:   sessions,
		now:        time.Now,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
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

	if err := s.accounts.UpsertAccount(accountID, playerID); err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}
	if err := s.sessions.SaveSession(session); err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}

	return pair, nil
}

// Refresh rotates the token pair for an existing refresh token.
func (s *AuthService) Refresh(refreshToken string) (domain.TokenPair, *apperrors.Error) {
	if refreshToken == "" {
		err := apperrors.New("invalid_request", "refresh_token is required", 400)
		return domain.TokenPair{}, &err
	}

	session, ok, err := s.sessions.GetSessionByRefreshToken(refreshToken)
	if err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}
	if !ok {
		err := apperrors.New("unauthorized", "refresh token is invalid", 401)
		return domain.TokenPair{}, &err
	}
	if !session.RefreshExpiresAt.IsZero() && session.RefreshExpiresAt.Before(s.now()) {
		_ = s.sessions.DeleteSessionByRefreshToken(refreshToken)
		err := apperrors.New("unauthorized", "refresh token has expired", 401)
		return domain.TokenPair{}, &err
	}

	if err := s.sessions.DeleteSessionByRefreshToken(refreshToken); err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}

	pair, newSession, err := s.issueTokens(session.AccountID, session.PlayerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}

	if err := s.accounts.UpsertAccount(newSession.AccountID, newSession.PlayerID); err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}
	if err := s.sessions.SaveSession(newSession); err != nil {
		internal := apperrors.Internal()
		return domain.TokenPair{}, &internal
	}
	return pair, nil
}

// Introspect resolves an access token into the authenticated subject.
func (s *AuthService) Introspect(accessToken string) (domain.Subject, *apperrors.Error) {
	if accessToken == "" {
		err := apperrors.New("invalid_request", "access token is required", 400)
		return domain.Subject{}, &err
	}

	session, ok, err := s.sessions.GetSessionByAccessToken(accessToken)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Subject{}, &internal
	}
	if !ok {
		err := apperrors.New("unauthorized", "access token is invalid", 401)
		return domain.Subject{}, &err
	}
	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Before(s.now()) {
		err := apperrors.New("unauthorized", "access token has expired", 401)
		return domain.Subject{}, &err
	}

	return domain.Subject{
		AccountID: session.AccountID,
		PlayerID:  session.PlayerID,
	}, nil
}

type lockedAccountStore struct {
	stores *memoryStores
}

func accountStoreWithLock(stores *memoryStores) *lockedAccountStore {
	return &lockedAccountStore{stores: stores}
}

func (s *lockedAccountStore) UpsertAccount(accountID string, playerID string) error {
	s.stores.mu.Lock()
	defer s.stores.mu.Unlock()
	return s.stores.accounts.UpsertAccount(accountID, playerID)
}

type lockedSessionStore struct {
	stores *memoryStores
}

func sessionStoreWithLock(stores *memoryStores) *lockedSessionStore {
	return &lockedSessionStore{stores: stores}
}

func (s *lockedSessionStore) SaveSession(session domain.Session) error {
	s.stores.mu.Lock()
	defer s.stores.mu.Unlock()
	return s.stores.sessions.SaveSession(session)
}

func (s *lockedSessionStore) GetSessionByRefreshToken(refreshToken string) (domain.Session, bool, error) {
	s.stores.mu.RLock()
	defer s.stores.mu.RUnlock()
	return s.stores.sessions.GetSessionByRefreshToken(refreshToken)
}

func (s *lockedSessionStore) GetSessionByAccessToken(accessToken string) (domain.Session, bool, error) {
	s.stores.mu.RLock()
	defer s.stores.mu.RUnlock()
	return s.stores.sessions.GetSessionByAccessToken(accessToken)
}

func (s *lockedSessionStore) DeleteSessionByRefreshToken(refreshToken string) error {
	s.stores.mu.Lock()
	defer s.stores.mu.Unlock()
	return s.stores.sessions.DeleteSessionByRefreshToken(refreshToken)
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
		AccessToken:      accessToken,
		AccountID:        accountID,
		PlayerID:         playerID,
		RefreshToken:     refreshToken,
		ExpiresAt:        s.now().Add(s.accessTTL),
		RefreshExpiresAt: s.now().Add(s.refreshTTL),
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    time.Duration(s.accessTTL.Seconds()),
		AccountID:    accountID,
		PlayerID:     playerID,
	}, session, nil
}
