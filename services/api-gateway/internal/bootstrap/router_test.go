package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUnknownPathNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/unknown", nil)
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
