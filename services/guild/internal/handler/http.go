package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	guildservice "github.com/xyun1996/social_backend/services/guild/internal/service"
)

// HTTPHandler exposes the early guild HTTP API.
type HTTPHandler struct {
	guilds *guildservice.GuildService
}

// NewHTTPHandler constructs the guild HTTP routes.
func NewHTTPHandler(guilds *guildservice.GuildService) *HTTPHandler {
	return &HTTPHandler{guilds: guilds}
}

type createGuildRequest struct {
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
}

type createInviteRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
	ToPlayerID    string `json:"to_player_id"`
}

type joinGuildRequest struct {
	InviteID      string `json:"invite_id"`
	ActorPlayerID string `json:"actor_player_id"`
}

type guildMemberActionRequest struct {
	ActorPlayerID  string `json:"actor_player_id"`
	TargetPlayerID string `json:"target_player_id"`
}

type updateAnnouncementRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
	Announcement  string `json:"announcement"`
}

// Routes returns the guild HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/guilds", h.handleCreateGuild)
	mux.HandleFunc("GET /v1/guilds/{guildID}", h.handleGetGuild)
	mux.HandleFunc("GET /v1/guilds/{guildID}/members", h.handleListMembers)
	mux.HandleFunc("POST /v1/guilds/{guildID}/invites", h.handleCreateInvite)
	mux.HandleFunc("POST /v1/guilds/{guildID}/join", h.handleJoinGuild)
	mux.HandleFunc("POST /v1/guilds/{guildID}/kick", h.handleKickMember)
	mux.HandleFunc("POST /v1/guilds/{guildID}/transfer-owner", h.handleTransferOwner)
	mux.HandleFunc("POST /v1/guilds/{guildID}/announcement", h.handleUpdateAnnouncement)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "guild",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleCreateGuild(w http.ResponseWriter, r *http.Request) {
	var request createGuildRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	guild, appErr := h.guilds.CreateGuild(request.Name, request.OwnerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, guild)
}

func (h *HTTPHandler) handleGetGuild(w http.ResponseWriter, r *http.Request) {
	guild, appErr := h.guilds.GetGuild(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, guild)
}

func (h *HTTPHandler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	members, appErr := h.guilds.ListMemberStates(r.Context(), r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"guild_id": r.PathValue("guildID"),
		"count":    len(members),
		"members":  members,
	})
}

func (h *HTTPHandler) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var request createInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	invite, appErr := h.guilds.CreateInvite(r.Context(), r.PathValue("guildID"), request.ActorPlayerID, request.ToPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, invite)
}

func (h *HTTPHandler) handleJoinGuild(w http.ResponseWriter, r *http.Request) {
	var request joinGuildRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	guild, appErr := h.guilds.JoinWithInvite(r.Context(), r.PathValue("guildID"), request.InviteID, request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, guild)
}

func (h *HTTPHandler) handleKickMember(w http.ResponseWriter, r *http.Request) {
	var request guildMemberActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	guild, appErr := h.guilds.KickMember(r.PathValue("guildID"), request.ActorPlayerID, request.TargetPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, guild)
}

func (h *HTTPHandler) handleTransferOwner(w http.ResponseWriter, r *http.Request) {
	var request guildMemberActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	guild, appErr := h.guilds.TransferOwnership(r.PathValue("guildID"), request.ActorPlayerID, request.TargetPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, guild)
}

func (h *HTTPHandler) handleUpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	var request updateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	guild, appErr := h.guilds.UpdateAnnouncement(r.PathValue("guildID"), request.ActorPlayerID, request.Announcement)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, guild)
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
