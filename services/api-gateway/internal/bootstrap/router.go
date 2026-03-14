package bootstrap

import (
	"net/http"

	"github.com/xyun1996/social_backend/pkg/transport"
	gatewayapp "github.com/xyun1996/social_backend/services/api-gateway/internal/app"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	runtime := gatewayapp.NewRuntime()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
			Service: "api-gateway",
			Status:  "ok",
		})
	})
	runtime.Mount(mux)

	return mux
}
