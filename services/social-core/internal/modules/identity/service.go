package identity

import (
	"fmt"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
)

const (
	accessTokenBytes  = 16
	refreshTokenBytes = 24
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in_seconds"`
	AccountID    string `json:"account_id"`
	PlayerID     string `json:"player_id"`
}

type Subject struct {
	AccountID string `json:"account_id"`
	PlayerID  string `json:"player_id"`
}

type session struct {
	AccessToken      string
	RefreshToken     string
	AccountID        string
	PlayerID         string
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
}

type Service struct {
	mu         sync.RWMutex
	now        func() time.Time
	accessTTL  time.Duration
	refreshTTL time.Duration
	byAccess   map[string]session
	byRefresh  map[string]session
}

func NewService(accessTTL, refreshTTL time.Duration) *Service {
	if accessTTL <= 0 {
		accessTTL = time.Hour
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	return &Service{
		now:        time.Now,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		byAccess:   make(map[string]session),
		byRefresh:  make(map[string]session),
	}
}

func (s *Service) Login(accountID, playerID string) (TokenPair, *apperrors.Error) {
	if accountID == "" || playerID == "" {
		err := apperrors.New("invalid_request", "account_id and player_id are required", 400)
		return TokenPair{}, &err
	}

	pair, record, appErr := s.issue(accountID, playerID)
	if appErr != nil {
		return TokenPair{}, appErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.byAccess[record.AccessToken] = record
	s.byRefresh[record.RefreshToken] = record
	return pair, nil
}

func (s *Service) Refresh(refreshToken string) (TokenPair, *apperrors.Error) {
	if refreshToken == "" {
		err := apperrors.New("invalid_request", "refresh_token is required", 400)
		return TokenPair{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.byRefresh[refreshToken]
	if !ok {
		err := apperrors.New("unauthorized", "refresh token is invalid", 401)
		return TokenPair{}, &err
	}
	if record.RefreshExpiresAt.Before(s.now()) {
		delete(s.byRefresh, refreshToken)
		delete(s.byAccess, record.AccessToken)
		err := apperrors.New("unauthorized", "refresh token has expired", 401)
		return TokenPair{}, &err
	}

	delete(s.byRefresh, refreshToken)
	delete(s.byAccess, record.AccessToken)

	pair, next, appErr := s.issue(record.AccountID, record.PlayerID)
	if appErr != nil {
		return TokenPair{}, appErr
	}

	s.byRefresh[next.RefreshToken] = next
	s.byAccess[next.AccessToken] = next
	return pair, nil
}

func (s *Service) Introspect(accessToken string) (Subject, *apperrors.Error) {
	if accessToken == "" {
		err := apperrors.New("invalid_request", "access token is required", 400)
		return Subject{}, &err
	}

	s.mu.RLock()
	record, ok := s.byAccess[accessToken]
	s.mu.RUnlock()
	if !ok {
		err := apperrors.New("unauthorized", "access token is invalid", 401)
		return Subject{}, &err
	}
	if record.ExpiresAt.Before(s.now()) {
		err := apperrors.New("unauthorized", "access token has expired", 401)
		return Subject{}, &err
	}

	return Subject{AccountID: record.AccountID, PlayerID: record.PlayerID}, nil
}

func (s *Service) issue(accountID, playerID string) (TokenPair, session, *apperrors.Error) {
	accessToken, err := idgen.Token(accessTokenBytes)
	if err != nil {
		internal := apperrors.Internal()
		return TokenPair{}, session{}, &internal
	}
	refreshToken, err := idgen.Token(refreshTokenBytes)
	if err != nil {
		internal := apperrors.Internal()
		return TokenPair{}, session{}, &internal
	}

	record := session{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccountID:        accountID,
		PlayerID:         playerID,
		ExpiresAt:        s.now().Add(s.accessTTL),
		RefreshExpiresAt: s.now().Add(s.refreshTTL),
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		AccountID:    accountID,
		PlayerID:     playerID,
	}, record, nil
}

func (s *Service) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("identity-service(sessions=%d)", len(s.byAccess))
}
