package bootstrap

import (
	"net/http"

	"github.com/xyun1996/social_backend/pkg/transport"
	coreapp "github.com/xyun1996/social_backend/services/social-core/internal/app"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	runtime := coreapp.NewRuntime()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
			Service: "social-core",
			Status:  "ok",
		})
	})
	runtime.MountRuntimeEndpoints(mux)

	return mux
}
