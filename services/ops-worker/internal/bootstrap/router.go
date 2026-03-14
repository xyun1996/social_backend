package bootstrap

import (
	"net/http"

	"github.com/xyun1996/social_backend/pkg/transport"
	opsapp "github.com/xyun1996/social_backend/services/ops-worker/internal/app"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	runtime := opsapp.NewRuntime()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
			Service: "ops-worker",
			Status:  "ok",
		})
	})
	runtime.Mount(mux)

	return mux
}
