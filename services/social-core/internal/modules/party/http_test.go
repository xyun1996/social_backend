package party

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	invitemodule "github.com/xyun1996/social_backend/services/social-core/internal/modules/invite"
)

func TestPartyCreateInviteJoinAndReady(t *testing.T) {
	invites := invitemodule.NewService()
	service := NewService(invites)
	handler := NewHTTPHandler(service)
	mux := http.NewServeMux()
	handler.Mount(mux)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/parties", bytes.NewReader([]byte(`{"leader_id":"leader-1"}`)))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("expected create party status %d, got %d", http.StatusOK, createRec.Code)
	}
	var party Party
	if err := json.Unmarshal(createRec.Body.Bytes(), &party); err != nil {
		t.Fatalf("decode party: %v", err)
	}

	createInviteReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+party.ID+"/invites", bytes.NewReader([]byte(`{"actor_player_id":"leader-1","to_player_id":"member-2"}`)))
	createInviteRec := httptest.NewRecorder()
	mux.ServeHTTP(createInviteRec, createInviteReq)
	if createInviteRec.Code != http.StatusOK {
		t.Fatalf("expected create invite status %d, got %d", http.StatusOK, createInviteRec.Code)
	}
	var invite invitemodule.Invite
	if err := json.Unmarshal(createInviteRec.Body.Bytes(), &invite); err != nil {
		t.Fatalf("decode invite: %v", err)
	}

	if _, appErr := invites.RespondInvite(invite.ID, "member-2", invitemodule.ActionAccept); appErr != nil {
		t.Fatalf("accept invite: %v", appErr)
	}

	joinReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+party.ID+"/join", bytes.NewReader([]byte(`{"invite_id":"`+invite.ID+`","actor_player_id":"member-2"}`)))
	joinRec := httptest.NewRecorder()
	mux.ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("expected join status %d, got %d", http.StatusOK, joinRec.Code)
	}

	readyReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+party.ID+"/ready", bytes.NewReader([]byte(`{"actor_player_id":"member-2","is_ready":true}`)))
	readyRec := httptest.NewRecorder()
	mux.ServeHTTP(readyRec, readyReq)
	if readyRec.Code != http.StatusOK {
		t.Fatalf("expected ready status %d, got %d", http.StatusOK, readyRec.Code)
	}

	membersReq := httptest.NewRequest(http.MethodGet, "/v1/parties/"+party.ID+"/members", nil)
	membersRec := httptest.NewRecorder()
	mux.ServeHTTP(membersRec, membersReq)
	if membersRec.Code != http.StatusOK {
		t.Fatalf("expected members status %d, got %d", http.StatusOK, membersRec.Code)
	}
	if !bytes.Contains(membersRec.Body.Bytes(), []byte("member-2")) {
		t.Fatalf("expected joined member in response, got %s", membersRec.Body.String())
	}
}
