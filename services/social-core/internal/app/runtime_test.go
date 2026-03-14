package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMountRuntimeEndpoints(t *testing.T) {
	runtime := NewRuntime()
	mux := http.NewServeMux()
	runtime.MountRuntimeEndpoints(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "guild-basics") {
		t.Fatalf("expected runtime payload to list registered modules, got %s", body)
	}
	if !strings.Contains(body, "\"authorizer\":false") {
		t.Fatalf("expected runtime payload to expose foundation readiness, got %s", body)
	}
}
