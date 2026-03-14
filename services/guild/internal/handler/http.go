package handler

import (
	"encoding/json"
	"net/http"
	"time"

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

type submitActivityRequest struct {
	ActorPlayerID  string `json:"actor_player_id"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	SourceType     string `json:"source_type,omitempty"`
}

// Routes returns the guild HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/guilds", h.handleCreateGuild)
	mux.HandleFunc("GET /v1/guilds/{guildID}", h.handleGetGuild)
	mux.HandleFunc("GET /v1/guild-memberships/{playerID}", h.handleFindGuildByPlayer)
	mux.HandleFunc("GET /v1/guilds/{guildID}/members", h.handleListMembers)
	mux.HandleFunc("GET /v1/guilds/{guildID}/logs", h.handleListLogs)
	mux.HandleFunc("GET /v1/guilds/{guildID}/progression", h.handleGetProgression)
	mux.HandleFunc("GET /v1/guilds/{guildID}/contributions", h.handleListContributions)
	mux.HandleFunc("GET /v1/guilds/{guildID}/rewards", h.handleListRewards)
	mux.HandleFunc("GET /v1/guilds/activity-templates", h.handleListActivityTemplates)
	mux.HandleFunc("GET /v1/guilds/{guildID}/activities", h.handleListActivities)
	mux.HandleFunc("GET /v1/guilds/{guildID}/activities/{templateKey}/instances", h.handleListActivityInstances)
	mux.HandleFunc("POST /v1/guilds/{guildID}/activities/{templateKey}", h.handleSubmitActivity)
	mux.HandleFunc("POST /v1/guilds/{guildID}/activities/{templateKey}/submit", h.handleSubmitActivity)
	mux.HandleFunc("POST /v1/internal/guilds/{guildID}/activities/ensure-current", h.handleEnsureCurrentActivityInstances)
	mux.HandleFunc("POST /v1/internal/guilds/{guildID}/activities/close-expired", h.handleCloseExpiredActivityInstances)
	mux.HandleFunc("POST /v1/guilds/{guildID}/invites", h.handleCreateInvite)
	mux.HandleFunc("POST /v1/guilds/{guildID}/join", h.handleJoinGuild)
	mux.HandleFunc("POST /v1/guilds/{guildID}/kick", h.handleKickMember)
	mux.HandleFunc("POST /v1/guilds/{guildID}/transfer-owner", h.handleTransferOwner)
	mux.HandleFunc("POST /v1/guilds/{guildID}/announcement", h.handleUpdateAnnouncement)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{Service: "guild", Status: "ok"})
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

func (h *HTTPHandler) handleFindGuildByPlayer(w http.ResponseWriter, r *http.Request) {
	guild, members, appErr := h.guilds.FindGuildByPlayer(r.Context(), r.PathValue("playerID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"id":                      guild.ID,
		"name":                    guild.Name,
		"owner_id":                guild.OwnerID,
		"announcement":            guild.Announcement,
		"announcement_updated_at": guild.AnnouncementUpdatedAt,
		"level":                   guild.Level,
		"experience":              guild.Experience,
		"count":                   len(members),
		"members":                 members,
	})
}

func (h *HTTPHandler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	members, appErr := h.guilds.ListMemberStates(r.Context(), r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(members), "members": members})
}

func (h *HTTPHandler) handleListLogs(w http.ResponseWriter, r *http.Request) {
	logs, appErr := h.guilds.ListLogs(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(logs), "logs": logs})
}

func (h *HTTPHandler) handleGetProgression(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.guilds.GetProgression(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListContributions(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.guilds.ListContributions(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(records), "contributions": records})
}

func (h *HTTPHandler) handleListRewards(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.guilds.ListRewardRecords(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(records), "rewards": records})
}

func (h *HTTPHandler) handleListActivityTemplates(w http.ResponseWriter, _ *http.Request) {
	templates := h.guilds.ListActivityTemplates()
	transport.WriteJSON(w, http.StatusOK, map[string]any{"count": len(templates), "templates": templates})
}

func (h *HTTPHandler) handleListActivities(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.guilds.ListActivities(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(records), "activities": records})
}

func (h *HTTPHandler) handleListActivityInstances(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.guilds.ListActivityInstances(r.PathValue("guildID"), r.PathValue("templateKey"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "template_key": r.PathValue("templateKey"), "count": len(records), "instances": records})
}

func (h *HTTPHandler) handleEnsureCurrentActivityInstances(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.guilds.EnsureCurrentActivityInstances(r.PathValue("guildID"), time.Now().UTC())
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(records), "instances": records})
}

func (h *HTTPHandler) handleCloseExpiredActivityInstances(w http.ResponseWriter, r *http.Request) {
	appErr := h.guilds.CloseExpiredActivityInstances(r.PathValue("guildID"), time.Now().UTC())
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "status": "ok"})
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

func (h *HTTPHandler) handleSubmitActivity(w http.ResponseWriter, r *http.Request) {
	var request submitActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, guild, progression, appErr := h.guilds.SubmitActivityWithOptions(r.Context(), r.PathValue("guildID"), request.ActorPlayerID, r.PathValue("templateKey"), request.IdempotencyKey, request.SourceType)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"record": record, "guild": guild, "progression": progression})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
