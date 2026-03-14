package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeEndpoints(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/runtime/status" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"runtime":"social-core","phase":"product-rebuild"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer upstream.Close()

	runtime := NewRuntime()
	runtime.APIGatewayURL = upstream.URL
	runtime.SocialCoreURL = upstream.URL

	mux := http.NewServeMux()
	runtime.Mount(mux)

	statusReq := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	statusRec := httptest.NewRecorder()
	mux.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected runtime status %d, got %d", http.StatusOK, statusRec.Code)
	}

	overviewReq := httptest.NewRequest(http.MethodGet, "/v1/support/product-overview", nil)
	overviewRec := httptest.NewRecorder()
	mux.ServeHTTP(overviewRec, overviewReq)
	if overviewRec.Code != http.StatusOK {
		t.Fatalf("expected overview status %d, got %d", http.StatusOK, overviewRec.Code)
	}

	repairReq := httptest.NewRequest(http.MethodPost, "/v1/support/repair/phase-a-sync", bytes.NewReader([]byte(`{"target":"player-1","dry_run":true}`)))
	repairRec := httptest.NewRecorder()
	mux.ServeHTTP(repairRec, repairReq)
	if repairRec.Code != http.StatusOK {
		t.Fatalf("expected repair status %d, got %d", http.StatusOK, repairRec.Code)
	}
}
