package handler

import (
	"encoding/json"
	"net/http"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/invite/internal/service"
)

// HTTPHandler exposes the early invite HTTP API.
type HTTPHandler struct {
	invites *service.InviteService
}

// NewHTTPHandler constructs the invite HTTP routes.
func NewHTTPHandler(invites *service.InviteService) *HTTPHandler {
	return &HTTPHandler{invites: invites}
}

type createInviteRequest struct {
	Domain       string `json:"domain"`
	ResourceID   string `json:"resource_id"`
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
	TTLSeconds   int    `json:"ttl_seconds"`
}

type respondInviteRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
}

// Routes returns the invite HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/invites", h.handleCreateInvite)
	mux.HandleFunc("GET /v1/invites/{inviteID}", h.handleGetInvite)
	mux.HandleFunc("POST /v1/invites/{inviteID}/accept", h.handleAcceptInvite)
	mux.HandleFunc("POST /v1/invites/{inviteID}/decline", h.handleDeclineInvite)
	mux.HandleFunc("GET /v1/invites", h.handleListInvites)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "invite",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var request createInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.invites.CreateInvite(
		request.Domain,
		request.ResourceID,
		request.FromPlayerID,
		request.ToPlayerID,
		time.Duration(request.TTLSeconds)*time.Second,
	)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handleGetInvite(w http.ResponseWriter, r *http.Request) {
	invite, appErr := h.invites.GetInvite(r.PathValue("inviteID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, invite)
}

func (h *HTTPHandler) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	h.handleRespondInvite(w, r, "accept")
}

func (h *HTTPHandler) handleDeclineInvite(w http.ResponseWriter, r *http.Request) {
	h.handleRespondInvite(w, r, "decline")
}

func (h *HTTPHandler) handleRespondInvite(w http.ResponseWriter, r *http.Request, action string) {
	var request respondInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.invites.RespondInvite(r.PathValue("inviteID"), request.ActorPlayerID, action)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handleListInvites(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")

	invites, appErr := h.invites.ListInvites(playerID, role, status)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"player_id": playerID,
		"role":      defaultString(role, "all"),
		"status":    status,
		"count":     len(invites),
		"invites":   invites,
	})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
