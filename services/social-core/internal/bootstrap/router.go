package bootstrap

import (
	"net/http"

	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/social-core/internal/modules"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
			Service: "social-core",
			Status:  "ok",
		})
	})
	mux.HandleFunc("GET /v1/runtime/status", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, map[string]any{
			"runtime": "social-core",
			"phase":   "product-rebuild",
			"modules": modules.ModuleNames,
		})
	})

	return mux
}
