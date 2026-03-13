package handler

import (
	"net/http"

	"github.com/xyun1996/social_backend/pkg/transport"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

// HTTPHandler exposes the early ops HTTP API.
type HTTPHandler struct {
	ops *opsservice.OpsService
}

// NewHTTPHandler constructs the ops HTTP routes.
func NewHTTPHandler(ops *opsservice.OpsService) *HTTPHandler {
	return &HTTPHandler{ops: ops}
}

// Routes returns the ops HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("GET /v1/ops/players/{playerID}/overview", h.handlePlayerOverview)
	mux.HandleFunc("GET /v1/ops/players/{playerID}/presence", h.handlePlayerPresence)
	mux.HandleFunc("GET /v1/ops/parties/{partyID}", h.handlePartySnapshot)
	mux.HandleFunc("GET /v1/ops/guilds/{guildID}", h.handleGuildSnapshot)
	mux.HandleFunc("GET /v1/ops/jobs", h.handleWorkerSnapshot)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "ops",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handlePlayerOverview(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.ops.GetPlayerOverview(r.Context(), r.PathValue("playerID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handlePlayerPresence(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.ops.GetPlayerPresence(r.Context(), r.PathValue("playerID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handlePartySnapshot(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.ops.GetPartySnapshot(r.Context(), r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleGuildSnapshot(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.ops.GetGuildSnapshot(r.Context(), r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleWorkerSnapshot(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.ops.GetWorkerSnapshot(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("type"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}
