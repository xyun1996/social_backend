package metrics

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type httpMetricKey struct {
	Method string `json:"method"`
	Route  string `json:"route"`
	Status int    `json:"status"`
}

type HTTPEndpointMetric struct {
	Method       string `json:"method"`
	Route        string `json:"route"`
	Status       int    `json:"status"`
	Count        uint64 `json:"count"`
	Bytes        uint64 `json:"bytes"`
	LatencyMSAvg uint64 `json:"latency_ms_avg"`
}

type httpMetricValue struct {
	count          uint64
	bytes          uint64
	totalLatencyMS uint64
}

// Registry stores a lightweight HTTP metrics snapshot for local production monitoring.
type Registry struct {
	service   string
	startedAt time.Time
	inflight  atomic.Int64
	mu        sync.RWMutex
	metrics   map[httpMetricKey]*httpMetricValue
}

// Snapshot is the JSON payload returned by the shared /metrics endpoint.
type Snapshot struct {
	Service   string               `json:"service"`
	StartedAt time.Time            `json:"started_at"`
	Inflight  int64                `json:"inflight"`
	Endpoints []HTTPEndpointMetric `json:"endpoints"`
}

// NewRegistry constructs an in-memory HTTP registry.
func NewRegistry(service string) *Registry {
	return &Registry{
		service:   service,
		startedAt: time.Now().UTC(),
		metrics:   make(map[httpMetricKey]*httpMetricValue),
	}
}

// IncInflight increments the active request count.
func (r *Registry) IncInflight() {
	if r == nil {
		return
	}

	r.inflight.Add(1)
}

// DecInflight decrements the active request count.
func (r *Registry) DecInflight() {
	if r == nil {
		return
	}

	r.inflight.Add(-1)
}

// Record tracks one completed HTTP request.
func (r *Registry) Record(method string, route string, status int, bytes int, duration time.Duration) {
	if r == nil {
		return
	}

	key := httpMetricKey{Method: method, Route: route, Status: status}

	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.metrics[key]
	if !ok {
		current = &httpMetricValue{}
		r.metrics[key] = current
	}

	current.count++
	if bytes > 0 {
		current.bytes += uint64(bytes)
	}
	current.totalLatencyMS += uint64(duration.Milliseconds())
}

// Snapshot returns a stable metrics view for observability endpoints.
func (r *Registry) Snapshot() Snapshot {
	if r == nil {
		return Snapshot{}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	endpoints := make([]HTTPEndpointMetric, 0, len(r.metrics))
	for key, value := range r.metrics {
		avg := uint64(0)
		if value.count > 0 {
			avg = value.totalLatencyMS / value.count
		}

		endpoints = append(endpoints, HTTPEndpointMetric{
			Method:       key.Method,
			Route:        key.Route,
			Status:       key.Status,
			Count:        value.count,
			Bytes:        value.bytes,
			LatencyMSAvg: avg,
		})
	}

	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Route != endpoints[j].Route {
			return endpoints[i].Route < endpoints[j].Route
		}
		if endpoints[i].Method != endpoints[j].Method {
			return endpoints[i].Method < endpoints[j].Method
		}
		return endpoints[i].Status < endpoints[j].Status
	})

	return Snapshot{
		Service:   r.service,
		StartedAt: r.startedAt,
		Inflight:  r.inflight.Load(),
		Endpoints: endpoints,
	}
}

// Handler exposes the registry through a shared JSON endpoint.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(r.Snapshot())
	})
}
