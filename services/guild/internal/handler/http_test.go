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

type fakePresenceReader struct {
	snapshots map[string]guildservice.PresenceSnapshot
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

func (f *fakePresenceReader) GetPresence(_ context.Context, playerID string) (guildservice.PresenceSnapshot, *apperrors.Error) {
	snapshot, ok := f.snapshots[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", http.StatusNotFound)
		return guildservice.PresenceSnapshot{}, &err
	}
	return snapshot, nil
}

func TestGuildLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{}
	presence := &fakePresenceReader{
		snapshots: map[string]guildservice.PresenceSnapshot{
			"p1": {PlayerID: "p1", Status: "online", SessionID: "sess-1"},
			"p2": {PlayerID: "p2", Status: "online", SessionID: "sess-2"},
		},
	}
	h := NewHTTPHandler(guildservice.NewGuildService(invites, presence))

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

	memberReq := httptest.NewRequest(http.MethodGet, "/v1/guilds/"+guildID+"/members", nil)
	memberRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(memberRec, memberReq)

	if memberRec.Code != http.StatusOK {
		t.Fatalf("unexpected members status: got %d want %d", memberRec.Code, http.StatusOK)
	}

	logReq := httptest.NewRequest(http.MethodGet, "/v1/guilds/"+guildID+"/logs", nil)
	logRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(logRec, logReq)
	if logRec.Code != http.StatusOK {
		t.Fatalf("unexpected logs status: got %d want %d", logRec.Code, http.StatusOK)
	}
}

func TestGuildManagementEndpoints(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{}
	presence := &fakePresenceReader{
		snapshots: map[string]guildservice.PresenceSnapshot{
			"p1": {PlayerID: "p1", Status: "online", SessionID: "sess-1"},
			"p2": {PlayerID: "p2", Status: "online", SessionID: "sess-2"},
		},
	}
	h := NewHTTPHandler(guildservice.NewGuildService(invites, presence))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/guilds", bytes.NewBufferString(`{"name":"Raiders","owner_id":"p1"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

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

	transferReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guildID+"/transfer-owner", bytes.NewBufferString(`{"actor_player_id":"p1","target_player_id":"p2"}`))
	transferRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(transferRec, transferReq)
	if transferRec.Code != http.StatusOK {
		t.Fatalf("unexpected transfer status: got %d want %d", transferRec.Code, http.StatusOK)
	}

	kickReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guildID+"/kick", bytes.NewBufferString(`{"actor_player_id":"p2","target_player_id":"p1"}`))
	kickRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(kickRec, kickReq)
	if kickRec.Code != http.StatusOK {
		t.Fatalf("unexpected kick status: got %d want %d", kickRec.Code, http.StatusOK)
	}
}

func TestGuildAnnouncementEndpoint(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{}
	presence := &fakePresenceReader{
		snapshots: map[string]guildservice.PresenceSnapshot{
			"p1": {PlayerID: "p1", Status: "online", SessionID: "sess-1"},
		},
	}
	h := NewHTTPHandler(guildservice.NewGuildService(invites, presence))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/guilds", bytes.NewBufferString(`{"name":"Raiders","owner_id":"p1"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	guildID, _ := created["id"].(string)

	updateReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guildID+"/announcement", bytes.NewBufferString(`{"actor_player_id":"p1","announcement":"Welcome to the guild"}`))
	updateRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("unexpected announcement update status: got %d want %d", updateRec.Code, http.StatusOK)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/guilds/"+guildID, nil)
	getRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected guild get status: got %d want %d", getRec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(getRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal get response: %v", err)
	}
	if payload["announcement"] != "Welcome to the guild" {
		t.Fatalf("unexpected announcement payload: %+v", payload)
	}
}

func TestGuildActivityEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(guildservice.NewGuildService(&fakeInviteClient{}, nil))

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

	templatesReq := httptest.NewRequest(http.MethodGet, "/v1/guilds/activity-templates", nil)
	templatesRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(templatesRec, templatesReq)
	if templatesRec.Code != http.StatusOK {
		t.Fatalf("unexpected templates status: got %d want %d", templatesRec.Code, http.StatusOK)
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/v1/guilds/"+guildID+"/activities/sign_in", bytes.NewBufferString(`{"actor_player_id":"p1"}`))
	submitRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(submitRec, submitReq)
	if submitRec.Code != http.StatusOK {
		t.Fatalf("unexpected activity submit status: got %d want %d", submitRec.Code, http.StatusOK)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/guilds/"+guildID+"/activities", nil)
	listRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected activity list status: got %d want %d", listRec.Code, http.StatusOK)
	}
}
