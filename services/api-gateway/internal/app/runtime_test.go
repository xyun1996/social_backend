package app

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestOwnsPath(t *testing.T) {
	tests := map[string]bool{
		"/v1/auth/login":                 true,
		"/v1/private-chat/conversations": true,
		"/v1/guilds":                     true,
		"/v1/parties/abc/members":        true,
		"/healthz":                       false,
		"/v1/runtime/status":             false,
		"/v1/unrelated":                  false,
	}

	for path, expected := range tests {
		if got := ownsPath(path); got != expected {
			t.Fatalf("ownsPath(%q) = %v, want %v", path, got, expected)
		}
	}
}

func TestRuntimeStatusAndProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/runtime/status":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"runtime":"social-core"}`))
		case "/v1/auth/login":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"proxied"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	runtime := NewRuntime()
	runtime.SocialCoreURL = backend.URL
	runtime.proxy = httputilProxy(backend.URL)

	mux := http.NewServeMux()
	runtime.Mount(mux)

	statusReq := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	statusRec := httptest.NewRecorder()
	mux.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected runtime status %d, got %d", http.StatusOK, statusRec.Code)
	}

	proxyReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	proxyRec := httptest.NewRecorder()
	mux.ServeHTTP(proxyRec, proxyReq)
	if proxyRec.Code != http.StatusOK {
		t.Fatalf("expected proxy status %d, got %d", http.StatusOK, proxyRec.Code)
	}
}

func httputilProxy(rawURL string) *httputil.ReverseProxy {
	target := mustParse(rawURL)
	return httputil.NewSingleHostReverseProxy(target)
}

func mustParse(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return parsed
}
