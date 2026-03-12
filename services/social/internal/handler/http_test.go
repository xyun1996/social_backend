package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xyun1996/social_backend/services/social/internal/service"
)

func TestFriendRequestLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewSocialService())

	sendReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewBufferString(`{"from_player_id":"p1","to_player_id":"p2"}`))
	sendRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(sendRec, sendReq)

	if sendRec.Code != http.StatusOK {
		t.Fatalf("unexpected send status: got %d want %d", sendRec.Code, http.StatusOK)
	}

	var requestPayload map[string]any
	if err := json.Unmarshal(sendRec.Body.Bytes(), &requestPayload); err != nil {
		t.Fatalf("unmarshal send response: %v", err)
	}

	requestID, _ := requestPayload["id"].(string)
	acceptReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests/"+requestID+"/accept", bytes.NewBufferString(`{"actor_player_id":"p2"}`))
	acceptRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(acceptRec, acceptReq)

	if acceptRec.Code != http.StatusOK {
		t.Fatalf("unexpected accept status: got %d want %d", acceptRec.Code, http.StatusOK)
	}
}

func TestBlockEndpointPreventsFutureFriendRequest(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewSocialService())

	blockReq := httptest.NewRequest(http.MethodPost, "/v1/blocks", bytes.NewBufferString(`{"player_id":"p2","blocked_player_id":"p1"}`))
	blockRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(blockRec, blockReq)

	if blockRec.Code != http.StatusOK {
		t.Fatalf("unexpected block status: got %d want %d", blockRec.Code, http.StatusOK)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewBufferString(`{"from_player_id":"p1","to_player_id":"p2"}`))
	sendRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(sendRec, sendReq)

	if sendRec.Code != http.StatusForbidden {
		t.Fatalf("unexpected send status after block: got %d want %d", sendRec.Code, http.StatusForbidden)
	}
}
