package bootstrap

import (
	"net/http"

	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/api-gateway/internal/modules"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
			Service: "api-gateway",
			Status:  "ok",
		})
	})
	mux.HandleFunc("GET /v1/runtime/status", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, map[string]any{
			"runtime": "api-gateway",
			"phase":   "product-rebuild",
			"modules": modules.ModuleNames,
			"target":  "client ingress boundary",
		})
	})

	return mux
}
