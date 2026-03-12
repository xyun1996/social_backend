package handler

import (
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

func TestSessionMeRequiresBearerToken(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(stubIntrospector{})
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
	})

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
