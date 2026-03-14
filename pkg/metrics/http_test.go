package metrics

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegistrySnapshot(t *testing.T) {
	t.Parallel()

	registry := NewRegistry("gateway")
	registry.IncInflight()
	registry.Record("GET", "/healthz", 200, 10, 25*time.Millisecond)
	registry.Record("GET", "/healthz", 200, 20, 35*time.Millisecond)

	snapshot := registry.Snapshot()
	if snapshot.Service != "gateway" {
		t.Fatalf("unexpected service: %q", snapshot.Service)
	}
	if snapshot.Inflight != 1 {
		t.Fatalf("unexpected inflight count: %d", snapshot.Inflight)
	}
	if len(snapshot.Endpoints) != 1 {
		t.Fatalf("unexpected endpoint count: %d", len(snapshot.Endpoints))
	}
	if snapshot.Endpoints[0].LatencyMSAvg != 30 {
		t.Fatalf("unexpected avg latency: %d", snapshot.Endpoints[0].LatencyMSAvg)
	}
}

func TestRegistryHandler(t *testing.T) {
	t.Parallel()

	registry := NewRegistry("chat")
	registry.Record("POST", "/v1/conversations", 200, 50, time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	registry.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	var payload Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if payload.Service != "chat" {
		t.Fatalf("unexpected service: %q", payload.Service)
	}
}
