package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xyun1996/social_backend/pkg/metrics"
)

func TestRequireInternalToken(t *testing.T) {
	t.Parallel()

	handler := RequireInternalToken("secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/internal/example", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status without token: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/internal/example", nil)
	req.Header.Set("X-Internal-Token", "secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status with token: %d", rec.Code)
	}
}

func TestRequireOpsToken(t *testing.T) {
	t.Parallel()

	handler := RequireOpsToken("ops-secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/ops/players/p1/overview", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status without token: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/ops/players/p1/overview", nil)
	req.Header.Set("Authorization", "Bearer ops-secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status with token: %d", rec.Code)
	}
}

func TestRateLimit(t *testing.T) {
	t.Parallel()

	handler := RateLimit(1, 0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected first status: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected second status: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected third status: %d", rec.Code)
	}
}

func TestAccessLogAddsRequestHeadersAndMetrics(t *testing.T) {
	t.Parallel()

	registry := metrics.NewRegistry("identity")
	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	handler := WithRequestContext(logger)(AccessLog(logger, registry)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("expected request id header")
	}

	payload := registry.Snapshot()
	if len(payload.Endpoints) != 1 {
		t.Fatalf("unexpected metric count: %d", len(payload.Endpoints))
	}
}

func TestRecoverReturnsInternalError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	handler := WithRequestContext(logger)(Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}
	if payload["code"] != "internal_error" {
		t.Fatalf("unexpected error payload: %+v", payload)
	}
}
