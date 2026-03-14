package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/chat/internal/service"
)

type fakePresenceReader struct {
	snapshots map[string]service.PresenceSnapshot
}

func (f *fakePresenceReader) GetPresence(_ context.Context, playerID string) (service.PresenceSnapshot, *apperrors.Error) {
	snapshot, ok := f.snapshots[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", http.StatusNotFound)
		return service.PresenceSnapshot{}, &err
	}

	return snapshot, nil
}

func TestChatLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewChatService(nil, nil))

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

	h := NewHTTPHandler(service.NewChatService(nil, nil))
	req := httptest.NewRequest(http.MethodGet, "/v1/conversations/c1/messages?player_id=p1&after_seq=bad", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected invalid query status: got %d want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDeliveryPlanEndpoint(t *testing.T) {
	t.Parallel()

	chat := service.NewChatService(&fakePresenceReader{
		snapshots: map[string]service.PresenceSnapshot{
			"p2": {
				PlayerID:  "p2",
				Status:    "online",
				SessionID: "sess-2",
			},
		},
	}, nil)
	h := NewHTTPHandler(chat)

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
	planReq := httptest.NewRequest(http.MethodGet, "/v1/conversations/"+conversationID+"/delivery?sender_player_id=p1", nil)
	planRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(planRec, planReq)

	if planRec.Code != http.StatusOK {
		t.Fatalf("unexpected delivery status: got %d want %d", planRec.Code, http.StatusOK)
	}
}

func TestChannelDescriptorEndpoint(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewChatService(nil, nil))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/conversations", bytes.NewBufferString(`{"kind":"guild","resource_id":"guild-1","member_player_ids":["p1","p2"]}`))
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

	channelReq := httptest.NewRequest(http.MethodGet, "/v1/conversations/"+conversationID+"/channel", nil)
	channelRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(channelRec, channelReq)
	if channelRec.Code != http.StatusOK {
		t.Fatalf("unexpected channel descriptor status: got %d want %d", channelRec.Code, http.StatusOK)
	}
}

func TestRecordOfflineDeliveryEndpoint(t *testing.T) {
	t.Parallel()

	chat := service.NewChatService(nil, nil)
	h := NewHTTPHandler(chat)

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

	var message map[string]any
	if err := json.Unmarshal(sendRec.Body.Bytes(), &message); err != nil {
		t.Fatalf("unmarshal send response: %v", err)
	}
	messageID, _ := message["id"].(string)

	recordReq := httptest.NewRequest(http.MethodPost, "/v1/internal/offline-deliveries", bytes.NewBufferString(`{"conversation_id":"`+conversationID+`","message_id":"`+messageID+`","recipient_player":"p2","delivery_mode":"offline_replay"}`))
	recordRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(recordRec, recordReq)
	if recordRec.Code != http.StatusOK {
		t.Fatalf("unexpected offline delivery status: got %d want %d", recordRec.Code, http.StatusOK)
	}
}
