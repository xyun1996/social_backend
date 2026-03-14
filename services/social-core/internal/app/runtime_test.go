package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMountRuntimeEndpoints(t *testing.T) {
	runtime := NewRuntime()
	mux := http.NewServeMux()
	runtime.MountRuntimeEndpoints(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "guild-basics") {
		t.Fatalf("expected runtime payload to list registered modules, got %s", body)
	}
	if !strings.Contains(body, "\"authorizer\":false") {
		t.Fatalf("expected runtime payload to expose foundation readiness, got %s", body)
	}
}

func TestMountedCorePhaseAEndpoints(t *testing.T) {
	runtime := NewRuntime()
	mux := http.NewServeMux()
	runtime.MountRuntimeEndpoints(mux)

	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader([]byte(`{"account_id":"account-1","player_id":"player-1"}`)))
	loginRec := httptest.NewRecorder()
	mux.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, loginRec.Code)
	}

	friendReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader([]byte(`{"from_player_id":"player-1","to_player_id":"player-2"}`)))
	friendRec := httptest.NewRecorder()
	mux.ServeHTTP(friendRec, friendReq)
	if friendRec.Code != http.StatusOK {
		t.Fatalf("expected friend request status %d, got %d", http.StatusOK, friendRec.Code)
	}
}
