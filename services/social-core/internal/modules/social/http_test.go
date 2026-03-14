package social

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFriendFlow(t *testing.T) {
	service := NewService()
	handler := NewHTTPHandler(service)
	mux := http.NewServeMux()
	handler.Mount(mux)

	sendReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader([]byte(`{"from_player_id":"p1","to_player_id":"p2"}`)))
	sendRec := httptest.NewRecorder()
	mux.ServeHTTP(sendRec, sendReq)
	if sendRec.Code != http.StatusOK {
		t.Fatalf("expected send status %d, got %d", http.StatusOK, sendRec.Code)
	}

	var request FriendRequest
	if err := json.Unmarshal(sendRec.Body.Bytes(), &request); err != nil {
		t.Fatalf("decode friend request: %v", err)
	}

	acceptReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests/"+request.ID+"/accept", bytes.NewReader([]byte(`{"actor_player_id":"p2"}`)))
	acceptRec := httptest.NewRecorder()
	mux.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected accept status %d, got %d", http.StatusOK, acceptRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/friends?player_id=p1", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, listRec.Code)
	}
	if !bytes.Contains(listRec.Body.Bytes(), []byte("p2")) {
		t.Fatalf("expected friend list to contain accepted friend, got %s", listRec.Body.String())
	}
}
