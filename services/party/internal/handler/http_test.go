package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	partyservice "github.com/xyun1996/social_backend/services/party/internal/service"
)

type fakeInviteClient struct {
	lastResourceID string
}

type fakePresenceReader struct {
	snapshots map[string]partyservice.PresenceSnapshot
}

func (f *fakeInviteClient) CreateInvite(_ context.Context, domainName string, resourceID string, fromPlayerID string, toPlayerID string) (partyservice.Invite, *apperrors.Error) {
	f.lastResourceID = resourceID
	return partyservice.Invite{
		ID:           "inv-1",
		Domain:       domainName,
		ResourceID:   resourceID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		Status:       "pending",
	}, nil
}

func (f *fakeInviteClient) GetInvite(_ context.Context, inviteID string) (partyservice.Invite, *apperrors.Error) {
	return partyservice.Invite{
		ID:           inviteID,
		Domain:       "party",
		ResourceID:   f.lastResourceID,
		FromPlayerID: "p1",
		ToPlayerID:   "p2",
		Status:       "accepted",
	}, nil
}

func (f *fakePresenceReader) GetPresence(_ context.Context, playerID string) (partyservice.PresenceSnapshot, *apperrors.Error) {
	snapshot, ok := f.snapshots[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", http.StatusNotFound)
		return partyservice.PresenceSnapshot{}, &err
	}
	return snapshot, nil
}

func TestPartyLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{}
	presence := &fakePresenceReader{
		snapshots: map[string]partyservice.PresenceSnapshot{
			"p1": {PlayerID: "p1", Status: "online", SessionID: "sess-1"},
			"p2": {PlayerID: "p2", Status: "online", SessionID: "sess-2"},
		},
	}
	svc := partyservice.NewPartyService(invites, presence)
	h := NewHTTPHandler(svc)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/parties", bytes.NewBufferString(`{"leader_id":"p1"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("unexpected create status: got %d want %d", createRec.Code, http.StatusOK)
	}

	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}

	partyID, _ := created["id"].(string)

	inviteReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/invites", bytes.NewBufferString(`{"actor_player_id":"p1","to_player_id":"p2"}`))
	inviteRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(inviteRec, inviteReq)

	if inviteRec.Code != http.StatusOK {
		t.Fatalf("unexpected invite status: got %d want %d", inviteRec.Code, http.StatusOK)
	}

	joinReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/join", bytes.NewBufferString(`{"invite_id":"inv-1","actor_player_id":"p2"}`))
	joinRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(joinRec, joinReq)

	if joinRec.Code != http.StatusOK {
		t.Fatalf("unexpected join status: got %d want %d", joinRec.Code, http.StatusOK)
	}

	readyReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/ready", bytes.NewBufferString(`{"actor_player_id":"p2","is_ready":true}`))
	readyRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(readyRec, readyReq)

	if readyRec.Code != http.StatusOK {
		t.Fatalf("unexpected ready status: got %d want %d", readyRec.Code, http.StatusOK)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/parties/"+partyID+"/ready", nil)
	listRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list ready status: got %d want %d", listRec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}

	if payload["count"].(float64) != 2 {
		t.Fatalf("unexpected ready count: %+v", payload)
	}

	memberReq := httptest.NewRequest(http.MethodGet, "/v1/parties/"+partyID+"/members", nil)
	memberRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(memberRec, memberReq)

	if memberRec.Code != http.StatusOK {
		t.Fatalf("unexpected members status: got %d want %d", memberRec.Code, http.StatusOK)
	}
}

func TestPartyManagementEndpoints(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{}
	presence := &fakePresenceReader{
		snapshots: map[string]partyservice.PresenceSnapshot{
			"p1": {PlayerID: "p1", Status: "online", SessionID: "sess-1"},
			"p2": {PlayerID: "p2", Status: "online", SessionID: "sess-2"},
		},
	}
	svc := partyservice.NewPartyService(invites, presence)
	h := NewHTTPHandler(svc)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/parties", bytes.NewBufferString(`{"leader_id":"p1"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	partyID, _ := created["id"].(string)

	inviteReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/invites", bytes.NewBufferString(`{"actor_player_id":"p1","to_player_id":"p2"}`))
	inviteRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(inviteRec, inviteReq)
	if inviteRec.Code != http.StatusOK {
		t.Fatalf("unexpected invite status: got %d want %d", inviteRec.Code, http.StatusOK)
	}

	joinReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/join", bytes.NewBufferString(`{"invite_id":"inv-1","actor_player_id":"p2"}`))
	joinRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("unexpected join status: got %d want %d", joinRec.Code, http.StatusOK)
	}

	transferReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/transfer-leader", bytes.NewBufferString(`{"actor_player_id":"p1","target_player_id":"p2"}`))
	transferRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(transferRec, transferReq)
	if transferRec.Code != http.StatusOK {
		t.Fatalf("unexpected transfer status: got %d want %d", transferRec.Code, http.StatusOK)
	}

	leaveReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/leave", bytes.NewBufferString(`{"actor_player_id":"p1"}`))
	leaveRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(leaveRec, leaveReq)
	if leaveRec.Code != http.StatusOK {
		t.Fatalf("unexpected leave status: got %d want %d", leaveRec.Code, http.StatusOK)
	}
}

func TestPartyQueueEndpoints(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{}
	presence := &fakePresenceReader{
		snapshots: map[string]partyservice.PresenceSnapshot{
			"p1": {PlayerID: "p1", Status: "online", SessionID: "sess-1"},
			"p2": {PlayerID: "p2", Status: "online", SessionID: "sess-2"},
		},
	}
	svc := partyservice.NewPartyService(invites, presence)
	h := NewHTTPHandler(svc)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/parties", bytes.NewBufferString(`{"leader_id":"p1"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	partyID, _ := created["id"].(string)

	inviteReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/invites", bytes.NewBufferString(`{"actor_player_id":"p1","to_player_id":"p2"}`))
	inviteRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(inviteRec, inviteReq)
	if inviteRec.Code != http.StatusOK {
		t.Fatalf("unexpected invite status: got %d want %d", inviteRec.Code, http.StatusOK)
	}

	joinReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/join", bytes.NewBufferString(`{"invite_id":"inv-1","actor_player_id":"p2"}`))
	joinRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("unexpected join status: got %d want %d", joinRec.Code, http.StatusOK)
	}

	for _, actor := range []string{"p1", "p2"} {
		readyReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/ready", bytes.NewBufferString(`{"actor_player_id":"`+actor+`","is_ready":true}`))
		readyRec := httptest.NewRecorder()
		h.Routes().ServeHTTP(readyRec, readyReq)
		if readyRec.Code != http.StatusOK {
			t.Fatalf("unexpected ready status for %s: got %d want %d", actor, readyRec.Code, http.StatusOK)
		}
	}

	queueJoinReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/queue/join", bytes.NewBufferString(`{"actor_player_id":"p1","queue_name":"casual-2v2"}`))
	queueJoinRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(queueJoinRec, queueJoinReq)
	if queueJoinRec.Code != http.StatusOK {
		t.Fatalf("unexpected queue join status: got %d want %d", queueJoinRec.Code, http.StatusOK)
	}

	queueGetReq := httptest.NewRequest(http.MethodGet, "/v1/parties/"+partyID+"/queue", nil)
	queueGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(queueGetRec, queueGetReq)
	if queueGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected queue get status: got %d want %d", queueGetRec.Code, http.StatusOK)
	}

	queueHandoffReq := httptest.NewRequest(http.MethodGet, "/v1/parties/"+partyID+"/queue/handoff", nil)
	queueHandoffRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(queueHandoffRec, queueHandoffReq)
	if queueHandoffRec.Code != http.StatusOK {
		t.Fatalf("unexpected queue handoff status: got %d want %d", queueHandoffRec.Code, http.StatusOK)
	}

	var handoff map[string]any
	if err := json.Unmarshal(queueHandoffRec.Body.Bytes(), &handoff); err != nil {
		t.Fatalf("unmarshal queue handoff response: %v", err)
	}
	ticketID, _ := handoff["ticket_id"].(string)

	assignReq := httptest.NewRequest(
		http.MethodPost,
		"/v1/parties/"+partyID+"/queue/assignment",
		bytes.NewBufferString(`{"ticket_id":"`+ticketID+`","match_id":"match-1","server_id":"game-1","connection_hint":"wss://game-1/session/match-1"}`),
	)
	assignRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(assignRec, assignReq)
	if assignRec.Code != http.StatusOK {
		t.Fatalf("unexpected queue assignment status: got %d want %d", assignRec.Code, http.StatusOK)
	}

	getAssignmentReq := httptest.NewRequest(http.MethodGet, "/v1/parties/"+partyID+"/queue/assignment", nil)
	getAssignmentRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(getAssignmentRec, getAssignmentReq)
	if getAssignmentRec.Code != http.StatusOK {
		t.Fatalf("unexpected queue assignment get status: got %d want %d", getAssignmentRec.Code, http.StatusOK)
	}

	queuedLeaveReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/leave", bytes.NewBufferString(`{"actor_player_id":"p2"}`))
	queuedLeaveRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(queuedLeaveRec, queuedLeaveReq)
	if queuedLeaveRec.Code != http.StatusConflict {
		t.Fatalf("expected queued party leave conflict, got %d", queuedLeaveRec.Code)
	}

	queueLeaveReq := httptest.NewRequest(http.MethodPost, "/v1/parties/"+partyID+"/queue/leave", bytes.NewBufferString(`{"actor_player_id":"p1"}`))
	queueLeaveRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(queueLeaveRec, queueLeaveReq)
	if queueLeaveRec.Code != http.StatusConflict {
		t.Fatalf("expected assigned queue leave conflict, got %d", queueLeaveRec.Code)
	}
}
