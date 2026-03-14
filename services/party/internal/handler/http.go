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

type leavePartyRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
}

type partyMemberActionRequest struct {
	ActorPlayerID  string `json:"actor_player_id"`
	TargetPlayerID string `json:"target_player_id"`
}

type queueJoinRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
	QueueName     string `json:"queue_name"`
}

type queueAssignmentRequest struct {
	TicketID       string `json:"ticket_id"`
	MatchID        string `json:"match_id"`
	ServerID       string `json:"server_id"`
	ConnectionHint string `json:"connection_hint"`
}

type queueResolutionRequest struct {
	TicketID string `json:"ticket_id"`
	MatchID  string `json:"match_id"`
	Status   string `json:"status"`
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
	mux.HandleFunc("POST /v1/parties/{partyID}/leave", h.handleLeaveParty)
	mux.HandleFunc("POST /v1/parties/{partyID}/kick", h.handleKickMember)
	mux.HandleFunc("POST /v1/parties/{partyID}/transfer-leader", h.handleTransferLeader)
	mux.HandleFunc("POST /v1/parties/{partyID}/queue/join", h.handleJoinQueue)
	mux.HandleFunc("POST /v1/parties/{partyID}/queue/leave", h.handleLeaveQueue)
	mux.HandleFunc("GET /v1/parties/{partyID}/queue", h.handleGetQueue)
	mux.HandleFunc("GET /v1/parties/{partyID}/queue/handoff", h.handleGetQueueHandoff)
	mux.HandleFunc("POST /v1/parties/{partyID}/queue/assignment", h.handleAssignMatch)
	mux.HandleFunc("GET /v1/parties/{partyID}/queue/assignment", h.handleGetQueueAssignment)
	mux.HandleFunc("POST /v1/parties/{partyID}/queue/assignment/resolve", h.handleResolveMatch)
	mux.HandleFunc("GET /v1/party-memberships/{playerID}", h.handleFindPartyByPlayer)
	mux.HandleFunc("GET /v1/parties/{partyID}/ready", h.handleListReady)
	mux.HandleFunc("GET /v1/parties/{partyID}/members", h.handleListMembers)
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

func (h *HTTPHandler) handleLeaveParty(w http.ResponseWriter, r *http.Request) {
	var request leavePartyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	party, appErr := h.parties.LeaveParty(r.PathValue("partyID"), request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, party)
}

func (h *HTTPHandler) handleKickMember(w http.ResponseWriter, r *http.Request) {
	var request partyMemberActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	party, appErr := h.parties.KickMember(r.PathValue("partyID"), request.ActorPlayerID, request.TargetPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, party)
}

func (h *HTTPHandler) handleTransferLeader(w http.ResponseWriter, r *http.Request) {
	var request partyMemberActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	party, appErr := h.parties.TransferLeader(r.PathValue("partyID"), request.ActorPlayerID, request.TargetPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, party)
}

func (h *HTTPHandler) handleJoinQueue(w http.ResponseWriter, r *http.Request) {
	var request queueJoinRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	state, appErr := h.parties.JoinQueue(r.Context(), r.PathValue("partyID"), request.ActorPlayerID, request.QueueName)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, state)
}

func (h *HTTPHandler) handleLeaveQueue(w http.ResponseWriter, r *http.Request) {
	var request leavePartyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.parties.LeaveQueue(r.PathValue("partyID"), request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handleGetQueue(w http.ResponseWriter, r *http.Request) {
	state, appErr := h.parties.GetQueueState(r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, state)
}

func (h *HTTPHandler) handleGetQueueHandoff(w http.ResponseWriter, r *http.Request) {
	handoff, members, appErr := h.parties.GetQueueHandoff(r.Context(), r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"ticket_id":    handoff.TicketID,
		"party_id":     handoff.PartyID,
		"queue_name":   handoff.QueueName,
		"leader_id":    handoff.LeaderID,
		"member_ids":   handoff.MemberIDs,
		"joined_at":    handoff.JoinedAt,
		"member_count": len(members),
		"members":      members,
	})
}

func (h *HTTPHandler) handleAssignMatch(w http.ResponseWriter, r *http.Request) {
	var request queueAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	assignment, appErr := h.parties.AssignMatch(
		r.Context(),
		r.PathValue("partyID"),
		request.TicketID,
		request.MatchID,
		request.ServerID,
		request.ConnectionHint,
	)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, assignment)
}

func (h *HTTPHandler) handleGetQueueAssignment(w http.ResponseWriter, r *http.Request) {
	assignment, appErr := h.parties.GetQueueAssignment(r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, assignment)
}

func (h *HTTPHandler) handleResolveMatch(w http.ResponseWriter, r *http.Request) {
	var request queueResolutionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	resolution, appErr := h.parties.ResolveMatch(r.PathValue("partyID"), request.TicketID, request.MatchID, request.Status)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, resolution)
}

func (h *HTTPHandler) handleFindPartyByPlayer(w http.ResponseWriter, r *http.Request) {
	party, members, appErr := h.parties.FindPartyByPlayer(r.Context(), r.PathValue("playerID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"party_id":   party.ID,
		"leader_id":  party.LeaderID,
		"member_ids": party.MemberIDs,
		"count":      len(members),
		"members":    members,
	})
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

func (h *HTTPHandler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	states, appErr := h.parties.ListMemberStates(r.Context(), r.PathValue("partyID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"party_id": r.PathValue("partyID"),
		"count":    len(states),
		"members":  states,
	})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
