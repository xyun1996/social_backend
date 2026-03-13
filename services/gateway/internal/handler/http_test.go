package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

type stubIntrospector struct {
	subject gatewayservice.Subject
	err     *apperrors.Error
}

func (s stubIntrospector) Introspect(context.Context, string) (gatewayservice.Subject, *apperrors.Error) {
	return s.subject, s.err
}

type stubPresenceReporter struct {
	snapshot gatewayservice.PresenceSnapshot
	err      *apperrors.Error
	update   gatewayservice.PresenceUpdate
}

func (s *stubPresenceReporter) Connect(_ context.Context, update gatewayservice.PresenceUpdate) (gatewayservice.PresenceSnapshot, *apperrors.Error) {
	s.update = update
	return s.snapshot, s.err
}

func (s *stubPresenceReporter) Heartbeat(_ context.Context, update gatewayservice.PresenceUpdate) (gatewayservice.PresenceSnapshot, *apperrors.Error) {
	s.update = update
	return s.snapshot, s.err
}

func (s *stubPresenceReporter) Disconnect(_ context.Context, update gatewayservice.PresenceUpdate) (gatewayservice.PresenceSnapshot, *apperrors.Error) {
	s.update = update
	return s.snapshot, s.err
}

type stubChatPlanner struct {
	targets             []gatewayservice.ChatDeliveryTarget
	messages            []gatewayservice.ChatReplayMessage
	err                 *apperrors.Error
	ackedConversationID string
	ackedPlayerID       string
	ackedSeq            int64
	replayedPlayerID    string
	replayedAfterSeq    int64
	replayedLimit       int
	replayedConvID      string
}

func (s stubChatPlanner) PlanDelivery(context.Context, string, string) ([]gatewayservice.ChatDeliveryTarget, *apperrors.Error) {
	return s.targets, s.err
}

func (s *stubChatPlanner) AckConversation(_ context.Context, conversationID string, playerID string, ackSeq int64) *apperrors.Error {
	s.ackedConversationID = conversationID
	s.ackedPlayerID = playerID
	s.ackedSeq = ackSeq
	return s.err
}

func (s *stubChatPlanner) ReplayMessages(_ context.Context, conversationID string, playerID string, afterSeq int64, limit int) ([]gatewayservice.ChatReplayMessage, *apperrors.Error) {
	s.replayedConvID = conversationID
	s.replayedPlayerID = playerID
	s.replayedAfterSeq = afterSeq
	s.replayedLimit = limit
	return s.messages, s.err
}

func TestSessionMeRequiresBearerToken(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(stubIntrospector{}, &stubPresenceReporter{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/session/me", nil)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSessionMeReturnsSubject(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(stubIntrospector{
		subject: gatewayservice.Subject{
			AccountID: "a1",
			PlayerID:  "p1",
		},
	}, &stubPresenceReporter{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/session/me", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["account_id"] != "a1" || payload["player_id"] != "p1" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestPresenceConnectReportsAuthenticatedPlayer(t *testing.T) {
	t.Parallel()

	reporter := &stubPresenceReporter{
		snapshot: gatewayservice.PresenceSnapshot{
			PlayerID:  "p1",
			Status:    "online",
			SessionID: "sess-1",
		},
	}

	h := NewHTTPHandler(stubIntrospector{
		subject: gatewayservice.Subject{
			AccountID: "a1",
			PlayerID:  "p1",
		},
	}, reporter, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/session/presence/connect", bytes.NewBufferString(`{"session_id":"sess-1","realm_id":"realm-1","location":"lobby"}`))
	req.Header.Set("Authorization", "Bearer token-1")
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}

	if reporter.update.PlayerID != "p1" || reporter.update.SessionID != "sess-1" {
		t.Fatalf("unexpected forwarded update: %+v", reporter.update)
	}
}

func TestRealtimeHandshakeAndResumeEndpoints(t *testing.T) {
	t.Parallel()

	reporter := &stubPresenceReporter{
		snapshot: gatewayservice.PresenceSnapshot{
			PlayerID:        "p1",
			Status:          "online",
			SessionID:       "sess-1",
			LastHeartbeatAt: "2026-03-13T10:00:00Z",
		},
	}
	h := NewHTTPHandler(stubIntrospector{
		subject: gatewayservice.Subject{AccountID: "a1", PlayerID: "p1"},
	}, reporter, nil)

	handshakeReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/handshake", bytes.NewBufferString(`{"access_token":"token-1","session_id":"sess-1","realm_id":"realm-1","location":"lobby","client_version":"dev"}`))
	handshakeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(handshakeRec, handshakeReq)
	if handshakeRec.Code != http.StatusOK {
		t.Fatalf("unexpected handshake status: got %d want %d", handshakeRec.Code, http.StatusOK)
	}

	resumeReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/resume", bytes.NewBufferString(`{"access_token":"token-1","session_id":"sess-1","last_server_event_id":"evt-42"}`))
	resumeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(resumeRec, resumeReq)
	if resumeRec.Code != http.StatusOK {
		t.Fatalf("unexpected resume status: got %d want %d", resumeRec.Code, http.StatusOK)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/realtime/sessions/sess-1", nil)
	getRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get session status: got %d want %d", getRec.Code, http.StatusOK)
	}
}

func TestRealtimeChatDeliveryEnqueuesSessionEvent(t *testing.T) {
	t.Parallel()

	reporter := &stubPresenceReporter{
		snapshot: gatewayservice.PresenceSnapshot{
			PlayerID:        "p2",
			Status:          "online",
			SessionID:       "sess-2",
			LastHeartbeatAt: "2026-03-13T10:00:00Z",
		},
	}
	h := NewHTTPHandler(stubIntrospector{
		subject: gatewayservice.Subject{AccountID: "a2", PlayerID: "p2"},
	}, reporter, &stubChatPlanner{
		targets: []gatewayservice.ChatDeliveryTarget{
			{PlayerID: "p2", DeliveryMode: "online_push", SessionID: "sess-2"},
			{PlayerID: "p3", DeliveryMode: "offline_replay"},
		},
	})

	handshakeReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/handshake", bytes.NewBufferString(`{"access_token":"token-2","session_id":"sess-2"}`))
	handshakeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(handshakeRec, handshakeReq)
	if handshakeRec.Code != http.StatusOK {
		t.Fatalf("unexpected handshake status: got %d want %d", handshakeRec.Code, http.StatusOK)
	}

	dispatchReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/chat/deliveries", bytes.NewBufferString(`{"conversation_id":"conv-1","sender_player_id":"p1","message_id":"msg-1","seq":1,"body":"hello","sent_at":"2026-03-13T10:00:00Z"}`))
	dispatchRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(dispatchRec, dispatchReq)
	if dispatchRec.Code != http.StatusOK {
		t.Fatalf("unexpected dispatch status: got %d want %d", dispatchRec.Code, http.StatusOK)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/realtime/sessions/sess-2/events", nil)
	eventsRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(eventsRec, eventsReq)
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("unexpected events status: got %d want %d", eventsRec.Code, http.StatusOK)
	}
}

func TestRealtimeChatAckUsesSessionPlayerIdentity(t *testing.T) {
	t.Parallel()

	reporter := &stubPresenceReporter{
		snapshot: gatewayservice.PresenceSnapshot{
			PlayerID:        "p2",
			Status:          "online",
			SessionID:       "sess-2",
			LastHeartbeatAt: "2026-03-13T10:00:00Z",
		},
	}
	planner := &stubChatPlanner{
		targets: []gatewayservice.ChatDeliveryTarget{
			{PlayerID: "p2", DeliveryMode: "online_push", SessionID: "sess-2"},
		},
	}
	h := NewHTTPHandler(stubIntrospector{
		subject: gatewayservice.Subject{AccountID: "a2", PlayerID: "p2"},
	}, reporter, planner)

	handshakeReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/handshake", bytes.NewBufferString(`{"access_token":"token-2","session_id":"sess-2"}`))
	handshakeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(handshakeRec, handshakeReq)
	if handshakeRec.Code != http.StatusOK {
		t.Fatalf("unexpected handshake status: got %d want %d", handshakeRec.Code, http.StatusOK)
	}

	dispatchReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/chat/deliveries", bytes.NewBufferString(`{"conversation_id":"conv-1","sender_player_id":"p1","message_id":"msg-1","seq":1,"body":"hello","sent_at":"2026-03-13T10:00:00Z"}`))
	dispatchRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(dispatchRec, dispatchReq)
	if dispatchRec.Code != http.StatusOK {
		t.Fatalf("unexpected dispatch status: got %d want %d", dispatchRec.Code, http.StatusOK)
	}

	ackReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/sessions/sess-2/acks", bytes.NewBufferString(`{"conversation_id":"conv-1","ack_seq":3}`))
	ackRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(ackRec, ackReq)
	if ackRec.Code != http.StatusOK {
		t.Fatalf("unexpected ack status: got %d want %d", ackRec.Code, http.StatusOK)
	}
	if planner.ackedConversationID != "conv-1" || planner.ackedPlayerID != "p2" || planner.ackedSeq != 3 {
		t.Fatalf("unexpected ack forwarding: %+v", planner)
	}

	var ackPayload map[string]any
	if err := json.Unmarshal(ackRec.Body.Bytes(), &ackPayload); err != nil {
		t.Fatalf("unmarshal ack response: %v", err)
	}
	if ackPayload["pruned_count"] != float64(1) {
		t.Fatalf("unexpected ack response payload: %+v", ackPayload)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/realtime/sessions/sess-2/events", nil)
	eventsRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(eventsRec, eventsReq)
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("unexpected events status: got %d want %d", eventsRec.Code, http.StatusOK)
	}

	var inbox struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(eventsRec.Body.Bytes(), &inbox); err != nil {
		t.Fatalf("unmarshal events response: %v", err)
	}
	if inbox.Count != 0 {
		t.Fatalf("expected session inbox to be compacted, got %+v", inbox)
	}
}

func TestRealtimeChatReplayUsesSessionPlayerIdentity(t *testing.T) {
	t.Parallel()

	reporter := &stubPresenceReporter{
		snapshot: gatewayservice.PresenceSnapshot{
			PlayerID:        "p2",
			Status:          "online",
			SessionID:       "sess-2",
			LastHeartbeatAt: "2026-03-13T10:00:00Z",
		},
	}
	planner := &stubChatPlanner{
		messages: []gatewayservice.ChatReplayMessage{
			{ID: "msg-2", ConversationID: "conv-1", Seq: 2, SenderPlayerID: "p1", Body: "world"},
		},
	}
	h := NewHTTPHandler(stubIntrospector{
		subject: gatewayservice.Subject{AccountID: "a2", PlayerID: "p2"},
	}, reporter, planner)

	handshakeReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/handshake", bytes.NewBufferString(`{"access_token":"token-2","session_id":"sess-2"}`))
	handshakeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(handshakeRec, handshakeReq)
	if handshakeRec.Code != http.StatusOK {
		t.Fatalf("unexpected handshake status: got %d want %d", handshakeRec.Code, http.StatusOK)
	}

	replayReq := httptest.NewRequest(http.MethodGet, "/v1/realtime/sessions/sess-2/replay?conversation_id=conv-1&after_seq=1&limit=20", nil)
	replayRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(replayRec, replayReq)
	if replayRec.Code != http.StatusOK {
		t.Fatalf("unexpected replay status: got %d want %d", replayRec.Code, http.StatusOK)
	}
	if planner.replayedConvID != "conv-1" || planner.replayedPlayerID != "p2" || planner.replayedAfterSeq != 1 || planner.replayedLimit != 20 {
		t.Fatalf("unexpected replay forwarding: %+v", planner)
	}
}

func TestRealtimeResumeTrimsBufferedEventsThroughLastSeenEvent(t *testing.T) {
	t.Parallel()

	reporter := &stubPresenceReporter{
		snapshot: gatewayservice.PresenceSnapshot{
			PlayerID:        "p2",
			Status:          "online",
			SessionID:       "sess-2",
			LastHeartbeatAt: "2026-03-13T10:00:00Z",
		},
	}
	h := NewHTTPHandler(stubIntrospector{
		subject: gatewayservice.Subject{AccountID: "a2", PlayerID: "p2"},
	}, reporter, &stubChatPlanner{
		targets: []gatewayservice.ChatDeliveryTarget{
			{PlayerID: "p2", DeliveryMode: "online_push", SessionID: "sess-2"},
		},
	})

	handshakeReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/handshake", bytes.NewBufferString(`{"access_token":"token-2","session_id":"sess-2"}`))
	handshakeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(handshakeRec, handshakeReq)
	if handshakeRec.Code != http.StatusOK {
		t.Fatalf("unexpected handshake status: got %d want %d", handshakeRec.Code, http.StatusOK)
	}

	for _, payload := range []string{
		`{"conversation_id":"conv-1","sender_player_id":"p1","message_id":"msg-1","seq":1,"body":"hello","sent_at":"2026-03-13T10:00:00Z"}`,
		`{"conversation_id":"conv-1","sender_player_id":"p1","message_id":"msg-2","seq":2,"body":"world","sent_at":"2026-03-13T10:00:01Z"}`,
	} {
		dispatchReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/chat/deliveries", bytes.NewBufferString(payload))
		dispatchRec := httptest.NewRecorder()
		h.Routes().ServeHTTP(dispatchRec, dispatchReq)
		if dispatchRec.Code != http.StatusOK {
			t.Fatalf("unexpected dispatch status: got %d want %d", dispatchRec.Code, http.StatusOK)
		}
	}

	resumeReq := httptest.NewRequest(http.MethodPost, "/v1/realtime/resume", bytes.NewBufferString(`{"access_token":"token-2","session_id":"sess-2","last_server_event_id":"msg-1:1"}`))
	resumeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(resumeRec, resumeReq)
	if resumeRec.Code != http.StatusOK {
		t.Fatalf("unexpected resume status: got %d want %d", resumeRec.Code, http.StatusOK)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/realtime/sessions/sess-2/events", nil)
	eventsRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(eventsRec, eventsReq)
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("unexpected events status: got %d want %d", eventsRec.Code, http.StatusOK)
	}

	var inbox struct {
		Count  int `json:"count"`
		Events []struct {
			EventID string `json:"event_id"`
		} `json:"events"`
	}
	if err := json.Unmarshal(eventsRec.Body.Bytes(), &inbox); err != nil {
		t.Fatalf("unmarshal events response: %v", err)
	}
	if inbox.Count != 1 || inbox.Events[0].EventID != "msg-2:2" {
		t.Fatalf("unexpected inbox after resume trim: %+v", inbox)
	}
}
