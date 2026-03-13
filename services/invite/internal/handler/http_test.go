package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xyun1996/social_backend/services/invite/internal/service"
)

func TestInviteLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewInviteService(nil))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/invites", bytes.NewBufferString(`{"domain":"party","resource_id":"party-1","from_player_id":"p1","to_player_id":"p2"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("unexpected create status: got %d want %d", createRec.Code, http.StatusOK)
	}

	var createPayload map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}

	inviteID, _ := createPayload["id"].(string)
	acceptReq := httptest.NewRequest(http.MethodPost, "/v1/invites/"+inviteID+"/accept", bytes.NewBufferString(`{"actor_player_id":"p2"}`))
	acceptRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(acceptRec, acceptReq)

	if acceptRec.Code != http.StatusOK {
		t.Fatalf("unexpected accept status: got %d want %d", acceptRec.Code, http.StatusOK)
	}
}

func TestListInvitesEndpoint(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewInviteService(nil))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/invites", bytes.NewBufferString(`{"domain":"guild","resource_id":"guild-1","from_player_id":"p1","to_player_id":"p2"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("unexpected create status: got %d want %d", createRec.Code, http.StatusOK)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/invites?player_id=p2&role=inbox&status=pending", nil)
	listRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: got %d want %d", listRec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}

	if payload["count"].(float64) != 1 {
		t.Fatalf("unexpected invite count: %+v", payload)
	}
}

func TestGetInviteEndpoint(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewInviteService(nil))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/invites", bytes.NewBufferString(`{"domain":"party","resource_id":"party-1","from_player_id":"p1","to_player_id":"p2"}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("unexpected create status: got %d want %d", createRec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}

	inviteID, _ := payload["id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/v1/invites/"+inviteID, nil)
	getRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: got %d want %d", getRec.Code, http.StatusOK)
	}
}
