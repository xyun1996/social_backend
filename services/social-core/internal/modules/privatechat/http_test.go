package privatechat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrivateChatFlow(t *testing.T) {
	service := NewService()
	handler := NewHTTPHandler(service)
	mux := http.NewServeMux()
	handler.Mount(mux)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/private-chat/conversations", bytes.NewReader([]byte(`{"member_player_ids":["p1","p2"]}`)))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("expected create status %d, got %d", http.StatusOK, createRec.Code)
	}

	var conversation Conversation
	if err := json.Unmarshal(createRec.Body.Bytes(), &conversation); err != nil {
		t.Fatalf("decode conversation: %v", err)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/v1/private-chat/conversations/"+conversation.ID+"/messages", bytes.NewReader([]byte(`{"sender_player_id":"p1","body":"hello"}`)))
	sendRec := httptest.NewRecorder()
	mux.ServeHTTP(sendRec, sendReq)
	if sendRec.Code != http.StatusOK {
		t.Fatalf("expected send status %d, got %d", http.StatusOK, sendRec.Code)
	}

	summaryReq := httptest.NewRequest(http.MethodGet, "/v1/private-chat/conversations/"+conversation.ID+"/summary?player_id=p2", nil)
	summaryRec := httptest.NewRecorder()
	mux.ServeHTTP(summaryRec, summaryReq)
	if summaryRec.Code != http.StatusOK {
		t.Fatalf("expected summary status %d, got %d", http.StatusOK, summaryRec.Code)
	}
	if !bytes.Contains(summaryRec.Body.Bytes(), []byte("\"unread_count\":1")) {
		t.Fatalf("expected unread summary, got %s", summaryRec.Body.String())
	}
}
