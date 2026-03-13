package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	partyservice "github.com/xyun1996/social_backend/services/party/internal/service"
)

// HTTPHandler exposes the early party HTTP API.
type HTTPHandler struct {
	parties *partyservice.PartyService
}

// NewHTTPHandler constructs the party HTTP routes.
func NewHTTPHandler(parties *partyservice.PartyService) *HTTPHandler {
	return &HTTPHandler{parties: parties}
}

type createPartyRequest struct {
	LeaderID string `json:"leader_id"`
}

type createInviteRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
	ToPlayerID    string `json:"to_player_id"`
}

type joinPartyRequest struct {
	InviteID      string `json:"invite_id"`
	ActorPlayerID string `json:"actor_player_id"`
}

type readyRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
	IsReady       bool   `json:"is_ready"`
}

// Routes returns the party HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/parties", h.handleCreateParty)
	mux.HandleFunc("GET /v1/parties/{partyID}", h.handleGetParty)
	mux.HandleFunc("POST /v1/parties/{partyID}/invites", h.handleCreateInvite)
	mux.HandleFunc("POST /v1/parties/{partyID}/join", h.handleJoinParty)
	mux.HandleFunc("POST /v1/parties/{partyID}/ready", h.handleSetReady)
	mux.HandleFunc("GET /v1/parties/{partyID}/ready", h.handleListReady)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "party",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleCreateParty(w http.ResponseWriter, r *http.Request) {
	var request createPartyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	party, appErr := h.parties.CreateParty(request.LeaderID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, party)
}

func (h *HTTPHandler) handleGetParty(w http.ResponseWriter, r *http.Request) {
	party, appErr := h.parties.GetParty(r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, party)
}

func (h *HTTPHandler) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var request createInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	invite, appErr := h.parties.CreateInvite(r.Context(), r.PathValue("partyID"), request.ActorPlayerID, request.ToPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, invite)
}

func (h *HTTPHandler) handleJoinParty(w http.ResponseWriter, r *http.Request) {
	var request joinPartyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	party, appErr := h.parties.JoinWithInvite(r.Context(), r.PathValue("partyID"), request.InviteID, request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, party)
}

func (h *HTTPHandler) handleSetReady(w http.ResponseWriter, r *http.Request) {
	var request readyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	state, appErr := h.parties.SetReady(r.PathValue("partyID"), request.ActorPlayerID, request.IsReady)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, state)
}

func (h *HTTPHandler) handleListReady(w http.ResponseWriter, r *http.Request) {
	states, appErr := h.parties.ListReadyStates(r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"party_id":     r.PathValue("partyID"),
		"count":        len(states),
		"ready_states": states,
	})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
