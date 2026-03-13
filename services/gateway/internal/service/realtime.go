package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

const (
	sessionStateActive = "active"
	sessionStateClosed = "closed"
)

// RealtimeSession is the gateway-owned runtime session view.
type RealtimeSession struct {
	SessionID         string `json:"session_id"`
	AccountID         string `json:"account_id"`
	PlayerID          string `json:"player_id"`
	RealmID           string `json:"realm_id,omitempty"`
	Location          string `json:"location,omitempty"`
	ClientVersion     string `json:"client_version,omitempty"`
	State             string `json:"state"`
	PresenceState     string `json:"presence_state"`
	LastServerEventID string `json:"last_server_event_id,omitempty"`
	ConnectedAt       string `json:"connected_at"`
	LastHeartbeatAt   string `json:"last_heartbeat_at,omitempty"`
	DisconnectedAt    string `json:"disconnected_at,omitempty"`
}

// HandshakeRequest is the HTTP prototype shape for realtime handshake.
type HandshakeRequest struct {
	AccessToken   string `json:"access_token"`
	SessionID     string `json:"session_id"`
	RealmID       string `json:"realm_id,omitempty"`
	Location      string `json:"location,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
}

// ResumeRequest is the HTTP prototype shape for realtime resume.
type ResumeRequest struct {
	AccessToken       string `json:"access_token"`
	SessionID         string `json:"session_id"`
	LastServerEventID string `json:"last_server_event_id,omitempty"`
}

// RealtimeService owns gateway-side realtime session lifecycle.
type RealtimeService struct {
	mu         sync.RWMutex
	sessions   map[string]RealtimeSession
	introspect Introspector
	presence   PresenceReporter
	now        func() time.Time
}

// NewRealtimeService constructs the realtime session prototype.
func NewRealtimeService(introspect Introspector, presence PresenceReporter) *RealtimeService {
	return &RealtimeService{
		sessions:   make(map[string]RealtimeSession),
		introspect: introspect,
		presence:   presence,
		now:        time.Now,
	}
}

// Handshake authenticates a new session and reports connect to presence.
func (s *RealtimeService) Handshake(ctx context.Context, request HandshakeRequest) (RealtimeSession, *apperrors.Error) {
	if request.AccessToken == "" || request.SessionID == "" {
		err := apperrors.New("invalid_request", "access_token and session_id are required", http.StatusBadRequest)
		return RealtimeSession{}, &err
	}

	if s.introspect == nil || s.presence == nil {
		internal := apperrors.New("dependency_missing", "gateway realtime dependencies are not configured", http.StatusInternalServerError)
		return RealtimeSession{}, &internal
	}

	subject, appErr := s.introspect.Introspect(ctx, request.AccessToken)
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	snapshot, appErr := s.presence.Connect(ctx, PresenceUpdate{
		PlayerID:  subject.PlayerID,
		SessionID: request.SessionID,
		RealmID:   request.RealmID,
		Location:  request.Location,
	})
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	now := s.now().UTC().Format(time.RFC3339Nano)
	session := RealtimeSession{
		SessionID:       request.SessionID,
		AccountID:       subject.AccountID,
		PlayerID:        subject.PlayerID,
		RealmID:         request.RealmID,
		Location:        request.Location,
		ClientVersion:   request.ClientVersion,
		State:           sessionStateActive,
		PresenceState:   snapshot.Status,
		ConnectedAt:     now,
		LastHeartbeatAt: snapshot.LastHeartbeatAt,
	}

	s.mu.Lock()
	s.sessions[session.SessionID] = session
	s.mu.Unlock()
	return session, nil
}

// Heartbeat refreshes presence for an active realtime session.
func (s *RealtimeService) Heartbeat(ctx context.Context, sessionID string) (RealtimeSession, *apperrors.Error) {
	session, appErr := s.getActiveSession(sessionID)
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	snapshot, appErr := s.presence.Heartbeat(ctx, PresenceUpdate{
		PlayerID:  session.PlayerID,
		SessionID: session.SessionID,
		RealmID:   session.RealmID,
		Location:  session.Location,
	})
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	s.mu.Lock()
	session.PresenceState = snapshot.Status
	session.LastHeartbeatAt = snapshot.LastHeartbeatAt
	s.sessions[sessionID] = session
	s.mu.Unlock()
	return session, nil
}

// Resume re-validates ownership and re-attaches runtime state to an existing session.
func (s *RealtimeService) Resume(ctx context.Context, request ResumeRequest) (RealtimeSession, *apperrors.Error) {
	if request.AccessToken == "" || request.SessionID == "" {
		err := apperrors.New("invalid_request", "access_token and session_id are required", http.StatusBadRequest)
		return RealtimeSession{}, &err
	}

	session, appErr := s.getSession(request.SessionID)
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	subject, appErr := s.introspect.Introspect(ctx, request.AccessToken)
	if appErr != nil {
		return RealtimeSession{}, appErr
	}
	if subject.PlayerID != session.PlayerID || subject.AccountID != session.AccountID {
		err := apperrors.New("forbidden", "session resume is only allowed for the original subject", http.StatusForbidden)
		return RealtimeSession{}, &err
	}

	snapshot, appErr := s.presence.Connect(ctx, PresenceUpdate{
		PlayerID:  session.PlayerID,
		SessionID: session.SessionID,
		RealmID:   session.RealmID,
		Location:  session.Location,
	})
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	s.mu.Lock()
	session.State = sessionStateActive
	session.PresenceState = snapshot.Status
	session.LastHeartbeatAt = snapshot.LastHeartbeatAt
	session.DisconnectedAt = ""
	session.LastServerEventID = request.LastServerEventID
	s.sessions[session.SessionID] = session
	s.mu.Unlock()
	return session, nil
}

// Close disconnects an active realtime session.
func (s *RealtimeService) Close(ctx context.Context, sessionID string) (RealtimeSession, *apperrors.Error) {
	session, appErr := s.getActiveSession(sessionID)
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	snapshot, appErr := s.presence.Disconnect(ctx, PresenceUpdate{
		PlayerID:  session.PlayerID,
		SessionID: session.SessionID,
		RealmID:   session.RealmID,
		Location:  session.Location,
	})
	if appErr != nil {
		return RealtimeSession{}, appErr
	}

	s.mu.Lock()
	session.State = sessionStateClosed
	session.PresenceState = snapshot.Status
	session.DisconnectedAt = s.now().UTC().Format(time.RFC3339Nano)
	s.sessions[session.SessionID] = session
	s.mu.Unlock()
	return session, nil
}

// GetSession returns the current stored gateway session view.
func (s *RealtimeService) GetSession(sessionID string) (RealtimeSession, *apperrors.Error) {
	return s.getSession(sessionID)
}

func (s *RealtimeService) getActiveSession(sessionID string) (RealtimeSession, *apperrors.Error) {
	session, appErr := s.getSession(sessionID)
	if appErr != nil {
		return RealtimeSession{}, appErr
	}
	if session.State != sessionStateActive {
		err := apperrors.New("invalid_state", "session is not active", http.StatusConflict)
		return RealtimeSession{}, &err
	}
	return session, nil
}

func (s *RealtimeService) getSession(sessionID string) (RealtimeSession, *apperrors.Error) {
	if sessionID == "" {
		err := apperrors.New("invalid_request", "session_id is required", http.StatusBadRequest)
		return RealtimeSession{}, &err
	}

	s.mu.RLock()
	session, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if !ok {
		err := apperrors.New("not_found", "session not found", http.StatusNotFound)
		return RealtimeSession{}, &err
	}
	return session, nil
}

func (s *RealtimeService) String() string {
	return fmt.Sprintf("gateway-realtime(sessions=%d)", len(s.sessions))
}
