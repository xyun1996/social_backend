package party

import (
	"encoding/json"
	"net/http"

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
	mux.HandleFunc("POST /v1/parties", h.handleCreateParty)
	mux.HandleFunc("GET /v1/parties/{partyID}", h.handleGetParty)
	mux.HandleFunc("GET /v1/party-memberships/{playerID}", h.handleFindPartyByPlayer)
	mux.HandleFunc("POST /v1/parties/{partyID}/invites", h.handleCreateInvite)
	mux.HandleFunc("POST /v1/parties/{partyID}/join", h.handleJoinParty)
	mux.HandleFunc("POST /v1/parties/{partyID}/ready", h.handleSetReady)
	mux.HandleFunc("GET /v1/parties/{partyID}/ready", h.handleListReady)
	mux.HandleFunc("GET /v1/parties/{partyID}/members", h.handleListMembers)
}

func (h *HTTPHandler) handleCreateParty(w http.ResponseWriter, r *http.Request) {
	var request struct {
		LeaderID string `json:"leader_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.CreateParty(request.LeaderID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleGetParty(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.service.GetParty(r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleFindPartyByPlayer(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.service.FindPartyByPlayer(r.PathValue("playerID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ActorPlayerID string `json:"actor_player_id"`
		ToPlayerID    string `json:"to_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.CreateInvite(r.PathValue("partyID"), request.ActorPlayerID, request.ToPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleJoinParty(w http.ResponseWriter, r *http.Request) {
	var request struct {
		InviteID      string `json:"invite_id"`
		ActorPlayerID string `json:"actor_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.JoinWithInvite(r.PathValue("partyID"), request.InviteID, request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleSetReady(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ActorPlayerID string `json:"actor_player_id"`
		IsReady       bool   `json:"is_ready"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.SetReady(r.PathValue("partyID"), request.ActorPlayerID, request.IsReady)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListReady(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.service.ListReadyStates(r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"party_id": r.PathValue("partyID"), "count": len(records), "ready_states": records})
}

func (h *HTTPHandler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.service.ListMemberStates(r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"party_id": r.PathValue("partyID"), "count": len(records), "members": records})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
