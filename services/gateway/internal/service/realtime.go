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

// SessionEventInbox is the gateway-owned buffered event list for a session.
type SessionEventInbox struct {
	SessionID string                `json:"session_id"`
	Count     int                   `json:"count"`
	Events    []ChatMessageEnvelope `json:"events"`
}

// AckCompactionResult summarizes local inbox pruning after a conversation ack.
type AckCompactionResult struct {
	SessionID        string `json:"session_id"`
	ConversationID   string `json:"conversation_id"`
	AckSeq           int64  `json:"ack_seq"`
	PrunedCount      int    `json:"pruned_count"`
	LastAckedEventID string `json:"last_acked_event_id,omitempty"`
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
	sessions   RealtimeSessionStore
	events     SessionEventStore
	introspect Introspector
	presence   PresenceReporter
	now        func() time.Time
}

// NewRealtimeService constructs the realtime session prototype.
func NewRealtimeService(introspect Introspector, presence PresenceReporter) *RealtimeService {
	return NewRealtimeServiceWithStores(newMemoryRealtimeSessionStore(), newMemorySessionEventStore(), introspect, presence)
}

// NewRealtimeServiceWithStores constructs the realtime service with injected persistence boundaries.
func NewRealtimeServiceWithStores(sessions RealtimeSessionStore, events SessionEventStore, introspect Introspector, presence PresenceReporter) *RealtimeService {
	return &RealtimeService{
		sessions:   sessions,
		events:     events,
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
	defer s.mu.Unlock()
	if err := s.sessions.SaveSession(session); err != nil {
		internal := apperrors.Internal()
		return RealtimeSession{}, &internal
	}
	events, err := s.events.GetEvents(session.SessionID)
	if err != nil {
		internal := apperrors.Internal()
		return RealtimeSession{}, &internal
	}
	if events == nil {
		if err := s.events.SaveEvents(session.SessionID, []ChatMessageEnvelope{}); err != nil {
			internal := apperrors.Internal()
			return RealtimeSession{}, &internal
		}
	}
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
	defer s.mu.Unlock()
	session.PresenceState = snapshot.Status
	session.LastHeartbeatAt = snapshot.LastHeartbeatAt
	if err := s.sessions.SaveSession(session); err != nil {
		internal := apperrors.Internal()
		return RealtimeSession{}, &internal
	}
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
	defer s.mu.Unlock()
	session.State = sessionStateActive
	session.PresenceState = snapshot.Status
	session.LastHeartbeatAt = snapshot.LastHeartbeatAt
	session.DisconnectedAt = ""
	session.LastServerEventID = request.LastServerEventID
	if request.LastServerEventID != "" {
		events, err := s.events.GetEvents(session.SessionID)
		if err != nil {
			internal := apperrors.Internal()
			return RealtimeSession{}, &internal
		}
		if err := s.events.SaveEvents(session.SessionID, trimEventsThroughID(events, request.LastServerEventID)); err != nil {
			internal := apperrors.Internal()
			return RealtimeSession{}, &internal
		}
	}
	if err := s.sessions.SaveSession(session); err != nil {
		internal := apperrors.Internal()
		return RealtimeSession{}, &internal
	}
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
	defer s.mu.Unlock()
	session.State = sessionStateClosed
	session.PresenceState = snapshot.Status
	session.DisconnectedAt = s.now().UTC().Format(time.RFC3339Nano)
	if err := s.sessions.SaveSession(session); err != nil {
		internal := apperrors.Internal()
		return RealtimeSession{}, &internal
	}
	return session, nil
}

// GetSession returns the current stored gateway session view.
func (s *RealtimeService) GetSession(sessionID string) (RealtimeSession, *apperrors.Error) {
	return s.getSession(sessionID)
}

// EnqueueChatEvent appends a chat event to an active session inbox.
func (s *RealtimeService) EnqueueChatEvent(sessionID string, event ChatMessageEnvelope) *apperrors.Error {
	session, appErr := s.getActiveSession(sessionID)
	if appErr != nil {
		return appErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	events, err := s.events.GetEvents(session.SessionID)
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	events = append(events, event)
	if err := s.events.SaveEvents(session.SessionID, events); err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	return nil
}

// GetSessionEvents returns the current event inbox for a session.
func (s *RealtimeService) GetSessionEvents(sessionID string) (SessionEventInbox, *apperrors.Error) {
	if _, appErr := s.getSession(sessionID); appErr != nil {
		return SessionEventInbox{}, appErr
	}

	events, err := s.events.GetEvents(sessionID)
	if err != nil {
		internal := apperrors.Internal()
		return SessionEventInbox{}, &internal
	}
	return SessionEventInbox{
		SessionID: sessionID,
		Count:     len(events),
		Events:    events,
	}, nil
}

// AcknowledgeConversation removes locally buffered chat events that are already durable on the client side.
func (s *RealtimeService) AcknowledgeConversation(sessionID string, conversationID string, ackSeq int64) (AckCompactionResult, *apperrors.Error) {
	session, appErr := s.getActiveSession(sessionID)
	if appErr != nil {
		return AckCompactionResult{}, appErr
	}
	if conversationID == "" {
		err := apperrors.New("invalid_request", "conversation_id is required", http.StatusBadRequest)
		return AckCompactionResult{}, &err
	}
	if ackSeq < 0 {
		err := apperrors.New("invalid_request", "ack_seq must be >= 0", http.StatusBadRequest)
		return AckCompactionResult{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	events, err := s.events.GetEvents(sessionID)
	if err != nil {
		internal := apperrors.Internal()
		return AckCompactionResult{}, &internal
	}
	kept := make([]ChatMessageEnvelope, 0, len(events))
	lastAckedEventID := ""
	prunedCount := 0
	for _, event := range events {
		if event.Stream == "chat" && event.Kind == "chat.message" && event.ConversationID == conversationID && event.Seq <= ackSeq {
			prunedCount++
			lastAckedEventID = event.EventID
			continue
		}
		kept = append(kept, event)
	}

	if err := s.events.SaveEvents(sessionID, kept); err != nil {
		internal := apperrors.Internal()
		return AckCompactionResult{}, &internal
	}
	if lastAckedEventID != "" {
		session.LastServerEventID = lastAckedEventID
		if err := s.sessions.SaveSession(session); err != nil {
			internal := apperrors.Internal()
			return AckCompactionResult{}, &internal
		}
	}

	return AckCompactionResult{
		SessionID:        sessionID,
		ConversationID:   conversationID,
		AckSeq:           ackSeq,
		PrunedCount:      prunedCount,
		LastAckedEventID: lastAckedEventID,
	}, nil
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
	session, ok, err := s.sessions.GetSession(sessionID)
	s.mu.RUnlock()
	if err != nil {
		internal := apperrors.Internal()
		return RealtimeSession{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "session not found", http.StatusNotFound)
		return RealtimeSession{}, &err
	}
	return session, nil
}

func (s *RealtimeService) String() string {
	sessions, err := s.sessions.ListSessions()
	if err != nil {
		return "gateway-realtime(sessions=unknown)"
	}
	return fmt.Sprintf("gateway-realtime(sessions=%d)", len(sessions))
}

func trimEventsThroughID(events []ChatMessageEnvelope, lastServerEventID string) []ChatMessageEnvelope {
	if lastServerEventID == "" {
		return events
	}

	trimIndex := -1
	for idx, event := range events {
		if event.EventID == lastServerEventID {
			trimIndex = idx
			break
		}
	}
	if trimIndex < 0 {
		return events
	}

	return append([]ChatMessageEnvelope(nil), events[trimIndex+1:]...)
}
