package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xyun1996/social_backend/services/presence/internal/service"
)

func TestPresenceLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewPresenceService())

	connectReq := httptest.NewRequest(http.MethodPost, "/v1/presence/connect", bytes.NewBufferString(`{"player_id":"p1","session_id":"sess-1","realm_id":"realm-1","location":"lobby"}`))
	connectRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(connectRec, connectReq)
	if connectRec.Code != http.StatusOK {
		t.Fatalf("unexpected connect status: got %d want %d", connectRec.Code, http.StatusOK)
	}

	heartbeatReq := httptest.NewRequest(http.MethodPost, "/v1/presence/heartbeat", bytes.NewBufferString(`{"player_id":"p1","session_id":"sess-1","realm_id":"realm-1","location":"queue"}`))
	heartbeatRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(heartbeatRec, heartbeatReq)
	if heartbeatRec.Code != http.StatusOK {
		t.Fatalf("unexpected heartbeat status: got %d want %d", heartbeatRec.Code, http.StatusOK)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/presence/p1", nil)
	getRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: got %d want %d", getRec.Code, http.StatusOK)
	}

	disconnectReq := httptest.NewRequest(http.MethodPost, "/v1/presence/disconnect", bytes.NewBufferString(`{"player_id":"p1","session_id":"sess-1"}`))
	disconnectRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(disconnectRec, disconnectReq)
	if disconnectRec.Code != http.StatusOK {
		t.Fatalf("unexpected disconnect status: got %d want %d", disconnectRec.Code, http.StatusOK)
	}
}
