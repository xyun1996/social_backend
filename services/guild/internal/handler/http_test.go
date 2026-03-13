package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	guildservice "github.com/xyun1996/social_backend/services/guild/internal/service"
)

type fakeInviteClient struct {
	lastResourceID string
}

func (f *fakeInviteClient) CreateInvite(_ context.Context, domainName string, resourceID string, fromPlayerID string, toPlayerID string) (guildservice.Invite, *apperrors.Error) {
	f.lastResourceID = resourceID
	return guildservice.Invite{
		ID:           "inv-1",
		Domain:       domainName,
		ResourceID:   resourceID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		Status:       "pending",
	}, nil
}

func (f *fakeInviteClient) GetInvite(_ context.Context, inviteID string) (guildservice.Invite, *apperrors.Error) {
	return guildservice.Invite{
		ID:           inviteID,
		Domain:       "guild",
		ResourceID:   f.lastResourceID,
		FromPlayerID: "p1",
		ToPlayerID:   "p2",
		Status:       "accepted",
	}, nil
}

func TestGuildLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{}
	h := NewHTTPHandler(guildservice.NewGuildService(invites))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/guilds", bytes.NewBufferString(`{"name":"Raiders","owner_id":"p1"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("unexpected create status: got %d want %d", createRec.Code, http.StatusOK)
	}

	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}

	guildID, _ := created["id"].(string)

	inviteReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guildID+"/invites", bytes.NewBufferString(`{"actor_player_id":"p1","to_player_id":"p2"}`))
	inviteRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(inviteRec, inviteReq)

	if inviteRec.Code != http.StatusOK {
		t.Fatalf("unexpected invite status: got %d want %d", inviteRec.Code, http.StatusOK)
	}

	joinReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guildID+"/join", bytes.NewBufferString(`{"invite_id":"inv-1","actor_player_id":"p2"}`))
	joinRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(joinRec, joinReq)

	if joinRec.Code != http.StatusOK {
		t.Fatalf("unexpected join status: got %d want %d", joinRec.Code, http.StatusOK)
	}
}
