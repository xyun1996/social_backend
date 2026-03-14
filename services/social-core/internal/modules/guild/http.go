package guild

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
	mux.HandleFunc("POST /v1/guilds", h.handleCreateGuild)
	mux.HandleFunc("GET /v1/guilds/{guildID}", h.handleGetGuild)
	mux.HandleFunc("GET /v1/guild-memberships/{playerID}", h.handleFindGuildByPlayer)
	mux.HandleFunc("GET /v1/guilds/{guildID}/members", h.handleListMembers)
	mux.HandleFunc("GET /v1/guilds/{guildID}/logs", h.handleListLogs)
	mux.HandleFunc("POST /v1/guilds/{guildID}/announcement", h.handleUpdateAnnouncement)
	mux.HandleFunc("POST /v1/guilds/{guildID}/invites", h.handleCreateInvite)
	mux.HandleFunc("POST /v1/guilds/{guildID}/join", h.handleJoinGuild)
}

func (h *HTTPHandler) handleCreateGuild(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name    string `json:"name"`
		OwnerID string `json:"owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.CreateGuild(request.Name, request.OwnerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleGetGuild(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.service.GetGuild(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleFindGuildByPlayer(w http.ResponseWriter, r *http.Request) {
	record, appErr := h.service.FindGuildByPlayer(r.PathValue("playerID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.service.ListMembers(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(records), "members": records})
}

func (h *HTTPHandler) handleListLogs(w http.ResponseWriter, r *http.Request) {
	records, appErr := h.service.ListLogs(r.PathValue("guildID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"guild_id": r.PathValue("guildID"), "count": len(records), "logs": records})
}

func (h *HTTPHandler) handleUpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ActorPlayerID string `json:"actor_player_id"`
		Announcement  string `json:"announcement"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.UpdateAnnouncement(r.PathValue("guildID"), request.ActorPlayerID, request.Announcement)
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
	record, appErr := h.service.CreateInvite(r.PathValue("guildID"), request.ActorPlayerID, request.ToPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleJoinGuild(w http.ResponseWriter, r *http.Request) {
	var request struct {
		InviteID      string `json:"invite_id"`
		ActorPlayerID string `json:"actor_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.JoinWithInvite(r.PathValue("guildID"), request.InviteID, request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
