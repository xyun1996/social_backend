package invite

import (
	"encoding/json"
	"net/http"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/invites", h.handleCreateInvite)
	mux.HandleFunc("GET /v1/invites", h.handleListInvites)
	mux.HandleFunc("GET /v1/invites/{inviteID}", h.handleGetInvite)
	mux.HandleFunc("POST /v1/invites/{inviteID}/accept", h.handleAcceptInvite)
	mux.HandleFunc("POST /v1/invites/{inviteID}/decline", h.handleDeclineInvite)
	mux.HandleFunc("POST /v1/invites/{inviteID}/cancel", h.handleCancelInvite)
	mux.HandleFunc("POST /v1/internal/invites/{inviteID}/expire", h.handleExpireInvite)
}

func (h *HTTPHandler) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Domain       string `json:"domain"`
		ResourceID   string `json:"resource_id"`
		FromPlayerID string `json:"from_player_id"`
		ToPlayerID   string `json:"to_player_id"`
		TTLSeconds   int    `json:"ttl_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.CreateInvite(
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
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListInvites(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")
	records, appErr := h.service.ListInvites(playerID, role, status)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	if role == "" {
		role = "all"
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"player_id": playerID,
		"role":      role,
		"status":    status,
		"count":     len(records),
		"invites":   records,
	})
}

func (h *HTTPHandler) handleGetInvite(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.service.GetInvite(r.PathValue("inviteID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	h.handleRespondInvite(w, r, ActionAccept)
}

func (h *HTTPHandler) handleDeclineInvite(w http.ResponseWriter, r *http.Request) {
	h.handleRespondInvite(w, r, ActionDecline)
}

func (h *HTTPHandler) handleRespondInvite(w http.ResponseWriter, r *http.Request, action string) {
	var request struct {
		ActorPlayerID string `json:"actor_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.RespondInvite(r.PathValue("inviteID"), request.ActorPlayerID, action)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleCancelInvite(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ActorPlayerID string `json:"actor_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.CancelInvite(r.PathValue("inviteID"), request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleExpireInvite(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.service.ExpireInvite(r.PathValue("inviteID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
