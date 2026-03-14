package social

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
	mux.HandleFunc("POST /v1/friends/requests", h.handleSendFriendRequest)
	mux.HandleFunc("GET /v1/friends/requests", h.handleListFriendRequests)
	mux.HandleFunc("POST /v1/friends/requests/{requestID}/accept", h.handleAcceptFriendRequest)
	mux.HandleFunc("GET /v1/friends", h.handleListFriends)
	mux.HandleFunc("POST /v1/friends/remarks", h.handleSetFriendRemark)
	mux.HandleFunc("GET /v1/friends/remarks", h.handleListFriendRemarks)
	mux.HandleFunc("POST /v1/blocks", h.handleBlockPlayer)
	mux.HandleFunc("GET /v1/blocks", h.handleListBlocks)
}

func (h *HTTPHandler) handleSendFriendRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		FromPlayerID string `json:"from_player_id"`
		ToPlayerID   string `json:"to_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.SendFriendRequest(request.FromPlayerID, request.ToPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListFriendRequests(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")
	records, appErr := h.service.ListFriendRequests(playerID, role, status)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"player_id": playerID,
		"role":      role,
		"status":    status,
		"count":     len(records),
		"requests":  records,
	})
}

func (h *HTTPHandler) handleAcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ActorPlayerID string `json:"actor_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.AcceptFriendRequest(r.PathValue("requestID"), request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListFriends(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	friends, appErr := h.service.ListFriends(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"player_id": playerID, "friends": friends})
}

func (h *HTTPHandler) handleSetFriendRemark(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID string `json:"player_id"`
		FriendID string `json:"friend_id"`
		Remark   string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.SetFriendRemark(request.PlayerID, request.FriendID, request.Remark)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListFriendRemarks(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	records, appErr := h.service.ListFriendRemarks(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"player_id": playerID, "count": len(records), "remarks": records})
}

func (h *HTTPHandler) handleBlockPlayer(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID        string `json:"player_id"`
		BlockedPlayerID string `json:"blocked_player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.BlockPlayer(request.PlayerID, request.BlockedPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListBlocks(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	records, appErr := h.service.ListBlocks(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"player_id": playerID, "blocks": records})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
