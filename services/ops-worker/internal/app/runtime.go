package app

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xyun1996/social_backend/pkg/transport"
)

type Runtime struct {
	APIGatewayURL string
	SocialCoreURL string
	client        *http.Client
}

func NewRuntime() Runtime {
	apiGatewayURL := strings.TrimSpace(os.Getenv("API_GATEWAY_BASE_URL"))
	if apiGatewayURL == "" {
		apiGatewayURL = "http://127.0.0.1:8090"
	}
	socialCoreURL := strings.TrimSpace(os.Getenv("SOCIAL_CORE_BASE_URL"))
	if socialCoreURL == "" {
		socialCoreURL = "http://127.0.0.1:8091"
	}

	return Runtime{
		APIGatewayURL: strings.TrimRight(apiGatewayURL, "/"),
		SocialCoreURL: strings.TrimRight(socialCoreURL, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (r Runtime) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/runtime/status", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, map[string]any{
			"runtime": "ops-worker",
			"phase":   "product-rebuild",
			"target": []string{
				"support-read-models",
				"repair-entrypoints",
				"release-readiness",
			},
			"upstreams": map[string]string{
				"api_gateway": r.APIGatewayURL,
				"social_core": r.SocialCoreURL,
			},
		})
	})

	mux.HandleFunc("GET /v1/support/product-overview", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, map[string]any{
			"phase":       "product-rebuild",
			"api_gateway": r.fetchRuntimeStatus(r.APIGatewayURL),
			"social_core": r.fetchRuntimeStatus(r.SocialCoreURL),
		})
	})

	mux.HandleFunc("POST /v1/support/repair/phase-a-sync", func(w http.ResponseWriter, req *http.Request) {
		var request struct {
			Target string `json:"target"`
			DryRun bool   `json:"dry_run"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			transport.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"code":    "invalid_json",
				"message": "request body must be valid json",
			})
			return
		}

		if strings.TrimSpace(request.Target) == "" {
			transport.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"code":    "invalid_request",
				"message": "target is required",
			})
			return
		}

		transport.WriteJSON(w, http.StatusOK, map[string]any{
			"status":   "accepted",
			"dry_run":  request.DryRun,
			"target":   request.Target,
			"runbook":  "phase-a-sync",
			"phase":    "product-rebuild",
			"upstream": r.fetchRuntimeStatus(r.SocialCoreURL),
		})
	})
}

func (r Runtime) fetchRuntimeStatus(baseURL string) map[string]any {
	result := map[string]any{
		"url":       baseURL,
		"reachable": false,
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/v1/runtime/status", nil)
	if err != nil {
		result["error"] = err.Error()
		return result
	}

	resp, err := r.client.Do(req)
	if err != nil {
		result["error"] = err.Error()
		return result
	}
	defer resp.Body.Close()

	result["status_code"] = resp.StatusCode
	result["reachable"] = resp.StatusCode == http.StatusOK

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
		result["payload"] = payload
	}

	return result
}
