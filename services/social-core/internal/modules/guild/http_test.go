package guild

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	invitemodule "github.com/xyun1996/social_backend/services/social-core/internal/modules/invite"
)

func TestGuildCreateInviteJoinAndAnnouncement(t *testing.T) {
	invites := invitemodule.NewService()
	service := NewService(invites)
	handler := NewHTTPHandler(service)
	mux := http.NewServeMux()
	handler.Mount(mux)

	createGuildReq := httptest.NewRequest(http.MethodPost, "/v1/guilds", bytes.NewReader([]byte(`{"name":"Guild One","owner_id":"owner-1"}`)))
	createGuildRec := httptest.NewRecorder()
	mux.ServeHTTP(createGuildRec, createGuildReq)
	if createGuildRec.Code != http.StatusOK {
		t.Fatalf("expected create guild status %d, got %d", http.StatusOK, createGuildRec.Code)
	}
	var guild Guild
	if err := json.Unmarshal(createGuildRec.Body.Bytes(), &guild); err != nil {
		t.Fatalf("decode guild: %v", err)
	}

	createInviteReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guild.ID+"/invites", bytes.NewReader([]byte(`{"actor_player_id":"owner-1","to_player_id":"member-2"}`)))
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

	joinReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guild.ID+"/join", bytes.NewReader([]byte(`{"invite_id":"`+invite.ID+`","actor_player_id":"member-2"}`)))
	joinRec := httptest.NewRecorder()
	mux.ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("expected join status %d, got %d", http.StatusOK, joinRec.Code)
	}

	announcementReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guild.ID+"/announcement", bytes.NewReader([]byte(`{"actor_player_id":"owner-1","announcement":"welcome"}`)))
	announcementRec := httptest.NewRecorder()
	mux.ServeHTTP(announcementRec, announcementReq)
	if announcementRec.Code != http.StatusOK {
		t.Fatalf("expected announcement status %d, got %d", http.StatusOK, announcementRec.Code)
	}

	membersReq := httptest.NewRequest(http.MethodGet, "/v1/guilds/"+guild.ID+"/members", nil)
	membersRec := httptest.NewRecorder()
	mux.ServeHTTP(membersRec, membersReq)
	if membersRec.Code != http.StatusOK {
		t.Fatalf("expected members status %d, got %d", http.StatusOK, membersRec.Code)
	}
	if !bytes.Contains(membersRec.Body.Bytes(), []byte("member-2")) {
		t.Fatalf("expected joined member in response, got %s", membersRec.Body.String())
	}
}
