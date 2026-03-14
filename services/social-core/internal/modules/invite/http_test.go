package invite

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInviteLifecycle(t *testing.T) {
	service := NewService()
	handler := NewHTTPHandler(service)
	mux := http.NewServeMux()
	handler.Mount(mux)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/invites", bytes.NewReader([]byte(`{"domain":"party","resource_id":"party-1","from_player_id":"p1","to_player_id":"p2"}`)))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("expected create status %d, got %d", http.StatusOK, createRec.Code)
	}

	var invite Invite
	if err := json.Unmarshal(createRec.Body.Bytes(), &invite); err != nil {
		t.Fatalf("decode invite: %v", err)
	}

	acceptReq := httptest.NewRequest(http.MethodPost, "/v1/invites/"+invite.ID+"/accept", bytes.NewReader([]byte(`{"actor_player_id":"p2"}`)))
	acceptRec := httptest.NewRecorder()
	mux.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected accept status %d, got %d", http.StatusOK, acceptRec.Code)
	}
	if !bytes.Contains(acceptRec.Body.Bytes(), []byte(StatusAccepted)) {
		t.Fatalf("expected accepted invite, got %s", acceptRec.Body.String())
	}
}
