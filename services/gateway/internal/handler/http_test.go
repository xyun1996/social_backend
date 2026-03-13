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
	targets []gatewayservice.ChatDeliveryTarget
	err     *apperrors.Error
}

func (s stubChatPlanner) PlanDelivery(context.Context, string, string) ([]gatewayservice.ChatDeliveryTarget, *apperrors.Error) {
	return s.targets, s.err
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
	}, reporter, stubChatPlanner{
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
