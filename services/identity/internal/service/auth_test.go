package service

import (
	"testing"
	"time"

	"github.com/xyun1996/social_backend/services/identity/internal/domain"
)

type recordingAccountStore struct {
	accountID string
	playerID  string
}

func (s *recordingAccountStore) UpsertAccount(accountID string, playerID string) error {
	s.accountID = accountID
	s.playerID = playerID
	return nil
}

type recordingSessionStore struct {
	sessionsByRefresh map[string]sessionRecord
	sessionsByAccess  map[string]sessionRecord
}

type sessionRecord struct {
	accountID        string
	playerID         string
	accessToken      string
	refreshToken     string
	expiresAt        time.Time
	refreshExpiresAt time.Time
}

func newRecordingSessionStore() *recordingSessionStore {
	return &recordingSessionStore{
		sessionsByRefresh: make(map[string]sessionRecord),
		sessionsByAccess:  make(map[string]sessionRecord),
	}
}

func (s *recordingSessionStore) SaveSession(session domain.Session) error {
	record := sessionRecord{
		accountID:        session.AccountID,
		playerID:         session.PlayerID,
		accessToken:      session.AccessToken,
		refreshToken:     session.RefreshToken,
		expiresAt:        session.ExpiresAt,
		refreshExpiresAt: session.RefreshExpiresAt,
	}
	s.sessionsByRefresh[session.RefreshToken] = record
	s.sessionsByAccess[session.AccessToken] = record
	return nil
}

func (s *recordingSessionStore) GetSessionByRefreshToken(refreshToken string) (domain.Session, bool, error) {
	record, ok := s.sessionsByRefresh[refreshToken]
	if !ok {
		return domain.Session{}, false, nil
	}
	return domain.Session{
		AccountID:        record.accountID,
		PlayerID:         record.playerID,
		AccessToken:      record.accessToken,
		RefreshToken:     record.refreshToken,
		ExpiresAt:        record.expiresAt,
		RefreshExpiresAt: record.refreshExpiresAt,
	}, true, nil
}

func (s *recordingSessionStore) GetSessionByAccessToken(accessToken string) (domain.Session, bool, error) {
	record, ok := s.sessionsByAccess[accessToken]
	if !ok {
		return domain.Session{}, false, nil
	}
	return domain.Session{
		AccountID:        record.accountID,
		PlayerID:         record.playerID,
		AccessToken:      record.accessToken,
		RefreshToken:     record.refreshToken,
		ExpiresAt:        record.expiresAt,
		RefreshExpiresAt: record.refreshExpiresAt,
	}, true, nil
}

func (s *recordingSessionStore) DeleteSessionByRefreshToken(refreshToken string) error {
	record, ok := s.sessionsByRefresh[refreshToken]
	if ok {
		delete(s.sessionsByRefresh, refreshToken)
		delete(s.sessionsByAccess, record.accessToken)
	}
	return nil
}

func TestLoginIssuesTokenPair(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()

	pair, err := svc.Login("account-1", "player-1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected tokens to be populated: %+v", pair)
	}

	if pair.TokenType != "Bearer" {
		t.Fatalf("unexpected token type: %q", pair.TokenType)
	}
}

func TestRefreshRotatesTokenPair(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()
	pair, err := svc.Login("account-1", "player-1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	refreshed, refreshErr := svc.Refresh(pair.RefreshToken)
	if refreshErr != nil {
		t.Fatalf("refresh returned error: %+v", refreshErr)
	}

	if refreshed.RefreshToken == pair.RefreshToken {
		t.Fatalf("expected refresh token rotation")
	}

	if _, secondErr := svc.Refresh(pair.RefreshToken); secondErr == nil {
		t.Fatalf("expected old refresh token to become invalid")
	}
}

func TestIntrospectReturnsSubjectForAccessToken(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()
	pair, err := svc.Login("account-1", "player-1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	subject, introspectErr := svc.Introspect(pair.AccessToken)
	if introspectErr != nil {
		t.Fatalf("introspect returned error: %+v", introspectErr)
	}

	if subject.AccountID != "account-1" || subject.PlayerID != "player-1" {
		t.Fatalf("unexpected subject: %+v", subject)
	}
}

func TestLoginUsesInjectedStores(t *testing.T) {
	t.Parallel()

	accounts := &recordingAccountStore{}
	sessions := newRecordingSessionStore()
	svc := NewAuthServiceWithStores(accounts, sessions)

	pair, err := svc.Login("account-9", "player-9")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	if accounts.accountID != "account-9" || accounts.playerID != "player-9" {
		t.Fatalf("unexpected stored account mapping: %+v", accounts)
	}
	if _, ok := sessions.sessionsByAccess[pair.AccessToken]; !ok {
		t.Fatalf("expected access token to be stored")
	}
	if _, ok := sessions.sessionsByRefresh[pair.RefreshToken]; !ok {
		t.Fatalf("expected refresh token to be stored")
	}
}

func TestRefreshRejectsExpiredRefreshToken(t *testing.T) {
	t.Parallel()

	accounts := &recordingAccountStore{}
	sessions := newRecordingSessionStore()
	svc := NewAuthServiceWithOptions(accounts, sessions, Options{
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Minute,
	})
	svc.now = func() time.Time { return time.Unix(1000, 0).UTC() }

	pair, err := svc.Login("account-1", "player-1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	svc.now = func() time.Time { return time.Unix(1000, 0).Add(2 * time.Minute).UTC() }
	if _, refreshErr := svc.Refresh(pair.RefreshToken); refreshErr == nil {
		t.Fatalf("expected expired refresh token to fail")
	}
}

func TestIntrospectRejectsExpiredAccessToken(t *testing.T) {
	t.Parallel()

	accounts := &recordingAccountStore{}
	sessions := newRecordingSessionStore()
	svc := NewAuthServiceWithOptions(accounts, sessions, Options{
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
	})
	svc.now = func() time.Time { return time.Unix(2000, 0).UTC() }

	pair, err := svc.Login("account-1", "player-1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	svc.now = func() time.Time { return time.Unix(2000, 0).Add(2 * time.Minute).UTC() }
	if _, introspectErr := svc.Introspect(pair.AccessToken); introspectErr == nil {
		t.Fatalf("expected expired access token to fail")
	}
}
