package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/presence/internal/service"
)

// HTTPHandler exposes the early presence HTTP API.
type HTTPHandler struct {
	presence *service.PresenceService
}

// NewHTTPHandler constructs the presence HTTP routes.
func NewHTTPHandler(presence *service.PresenceService) *HTTPHandler {
	return &HTTPHandler{presence: presence}
}

type updatePresenceRequest struct {
	PlayerID  string `json:"player_id"`
	SessionID string `json:"session_id"`
	RealmID   string `json:"realm_id"`
	Location  string `json:"location"`
}

type disconnectRequest struct {
	PlayerID  string `json:"player_id"`
	SessionID string `json:"session_id"`
}

// Routes returns the presence HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/presence/connect", h.handleConnect)
	mux.HandleFunc("POST /v1/presence/heartbeat", h.handleHeartbeat)
	mux.HandleFunc("POST /v1/presence/disconnect", h.handleDisconnect)
	mux.HandleFunc("GET /v1/presence/{playerID}", h.handleGetPresence)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "presence",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleConnect(w http.ResponseWriter, r *http.Request) {
	var request updatePresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	presence, appErr := h.presence.Connect(request.PlayerID, request.SessionID, request.RealmID, request.Location)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, presence)
}

func (h *HTTPHandler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var request updatePresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	presence, appErr := h.presence.Heartbeat(request.PlayerID, request.SessionID, request.RealmID, request.Location)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, presence)
}

func (h *HTTPHandler) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	var request disconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	presence, appErr := h.presence.Disconnect(request.PlayerID, request.SessionID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, presence)
}

func (h *HTTPHandler) handleGetPresence(w http.ResponseWriter, r *http.Request) {
	presence, appErr := h.presence.GetPresence(r.PathValue("playerID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, presence)
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
