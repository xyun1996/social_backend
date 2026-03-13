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

func TestSessionMeRequiresBearerToken(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(stubIntrospector{}, &stubPresenceReporter{})
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
	}, &stubPresenceReporter{})

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
	}, reporter)

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
