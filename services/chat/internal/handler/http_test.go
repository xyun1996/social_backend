package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xyun1996/social_backend/services/chat/internal/service"
)

func TestChatLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewChatService())

	createReq := httptest.NewRequest(http.MethodPost, "/v1/conversations", bytes.NewBufferString(`{"kind":"private","member_player_ids":["p1","p2"]}`))
	createRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("unexpected create status: got %d want %d", createRec.Code, http.StatusOK)
	}

	var conversation map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &conversation); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}

	conversationID, _ := conversation["id"].(string)

	sendReq := httptest.NewRequest(http.MethodPost, "/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"sender_player_id":"p1","body":"hello"}`))
	sendRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(sendRec, sendReq)

	if sendRec.Code != http.StatusOK {
		t.Fatalf("unexpected send status: got %d want %d", sendRec.Code, http.StatusOK)
	}

	ackReq := httptest.NewRequest(http.MethodPost, "/v1/conversations/"+conversationID+"/ack", bytes.NewBufferString(`{"player_id":"p2","ack_seq":1}`))
	ackRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(ackRec, ackReq)

	if ackRec.Code != http.StatusOK {
		t.Fatalf("unexpected ack status: got %d want %d", ackRec.Code, http.StatusOK)
	}

	replayReq := httptest.NewRequest(http.MethodGet, "/v1/conversations/"+conversationID+"/messages?player_id=p2&after_seq=0", nil)
	replayRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(replayRec, replayReq)

	if replayRec.Code != http.StatusOK {
		t.Fatalf("unexpected replay status: got %d want %d", replayRec.Code, http.StatusOK)
	}
}

func TestReplayRejectsInvalidQuery(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewChatService())
	req := httptest.NewRequest(http.MethodGet, "/v1/conversations/c1/messages?player_id=p1&after_seq=bad", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected invalid query status: got %d want %d", rec.Code, http.StatusBadRequest)
	}
}
