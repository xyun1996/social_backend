package app

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xyun1996/social_backend/pkg/transport"
)

var phaseAPathPrefixes = []string{
	"/v1/auth/",
	"/v1/friends",
	"/v1/blocks",
	"/v1/invites",
	"/v1/private-chat/",
	"/v1/guilds",
	"/v1/guild-memberships/",
	"/v1/parties",
	"/v1/party-memberships/",
}

type Runtime struct {
	SocialCoreURL string
	client        *http.Client
	proxy         *httputil.ReverseProxy
}

func NewRuntime() Runtime {
	baseURL := strings.TrimSpace(os.Getenv("SOCIAL_CORE_BASE_URL"))
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8091"
	}
	target, err := url.Parse(baseURL)
	if err != nil {
		target, _ = url.Parse("http://127.0.0.1:8091")
		baseURL = target.String()
	}

	return Runtime{
		SocialCoreURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		proxy: httputil.NewSingleHostReverseProxy(target),
	}
}

func (r Runtime) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/runtime/status", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, map[string]any{
			"runtime":      "api-gateway",
			"phase":        "product-rebuild",
			"upstream":     "social-core",
			"upstream_url": r.SocialCoreURL,
			"paths":        phaseAPathPrefixes,
		})
	})
	mux.HandleFunc("GET /v1/runtime/upstreams", func(w http.ResponseWriter, _ *http.Request) {
		status := map[string]any{
			"social_core_url": r.SocialCoreURL,
			"reachable":       false,
		}

		req, err := http.NewRequest(http.MethodGet, r.SocialCoreURL+"/v1/runtime/status", nil)
		if err == nil {
			resp, err := r.client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				status["reachable"] = resp.StatusCode == http.StatusOK
				status["status_code"] = resp.StatusCode
				var payload map[string]any
				if decodeErr := json.NewDecoder(resp.Body).Decode(&payload); decodeErr == nil {
					status["payload"] = payload
				}
			} else {
				status["error"] = err.Error()
			}
		} else {
			status["error"] = err.Error()
		}

		transport.WriteJSON(w, http.StatusOK, status)
	})
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if ownsPath(req.URL.Path) {
			r.proxy.ServeHTTP(w, req)
			return
		}
		http.NotFound(w, req)
	}))
}

func ownsPath(path string) bool {
	for _, prefix := range phaseAPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
