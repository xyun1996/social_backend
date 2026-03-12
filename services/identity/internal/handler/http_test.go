package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xyun1996/social_backend/services/identity/internal/service"
)

func TestLoginEndpoint(t *testing.T) {
	t.Parallel()

	h := NewAuthHTTPHandler(service.NewAuthService())
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"account_id":"a1","player_id":"p1"}`))
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

func TestRefreshEndpointRequiresToken(t *testing.T) {
	t.Parallel()

	h := NewAuthHTTPHandler(service.NewAuthService())
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBufferString(`{"refresh_token":""}`))
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestIntrospectEndpointReturnsSubject(t *testing.T) {
	t.Parallel()

	auth := service.NewAuthService()
	pair, err := auth.Login("a1", "p1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	h := NewAuthHTTPHandler(auth)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/introspect", bytes.NewBufferString(`{"access_token":"`+pair.AccessToken+`"}`))
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
