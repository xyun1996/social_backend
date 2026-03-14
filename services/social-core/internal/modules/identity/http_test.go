package identity

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginAndIntrospect(t *testing.T) {
	service := NewService(0, 0)
	handler := NewHTTPHandler(service)
	mux := http.NewServeMux()
	handler.Mount(mux)

	loginBody := []byte(`{"account_id":"account-1","player_id":"player-1"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(loginBody))
	loginRec := httptest.NewRecorder()
	mux.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, loginRec.Code)
	}

	var pair TokenPair
	if err := json.Unmarshal(loginRec.Body.Bytes(), &pair); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	introspectBody := []byte(`{"access_token":"` + pair.AccessToken + `"}`)
	introspectReq := httptest.NewRequest(http.MethodPost, "/v1/auth/introspect", bytes.NewReader(introspectBody))
	introspectRec := httptest.NewRecorder()
	mux.ServeHTTP(introspectRec, introspectReq)

	if introspectRec.Code != http.StatusOK {
		t.Fatalf("expected introspect status %d, got %d", http.StatusOK, introspectRec.Code)
	}
}
