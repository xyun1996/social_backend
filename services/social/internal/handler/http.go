package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/social/internal/service"
)

// HTTPHandler exposes the early social HTTP API.
type HTTPHandler struct {
	social *service.SocialService
}

// NewHTTPHandler constructs the social HTTP routes.
func NewHTTPHandler(social *service.SocialService) *HTTPHandler {
	return &HTTPHandler{social: social}
}

type sendFriendRequest struct {
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
}

type acceptFriendRequest struct {
	ActorPlayerID string `json:"actor_player_id"`
}

type blockRequest struct {
	PlayerID        string `json:"player_id"`
	BlockedPlayerID string `json:"blocked_player_id"`
}

// Routes returns the social HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/friends/requests", h.handleSendFriendRequest)
	mux.HandleFunc("POST /v1/friends/requests/{requestID}/accept", h.handleAcceptFriendRequest)
	mux.HandleFunc("GET /v1/friends", h.handleListFriends)
	mux.HandleFunc("POST /v1/blocks", h.handleBlockPlayer)
	mux.HandleFunc("GET /v1/blocks", h.handleListBlocks)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "social",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleSendFriendRequest(w http.ResponseWriter, r *http.Request) {
	var request sendFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.social.SendFriendRequest(request.FromPlayerID, request.ToPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handleAcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	var request acceptFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.social.AcceptFriendRequest(r.PathValue("requestID"), request.ActorPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handleListFriends(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	friends, appErr := h.social.ListFriends(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"player_id": playerID,
		"friends":   friends,
	})
}

func (h *HTTPHandler) handleBlockPlayer(w http.ResponseWriter, r *http.Request) {
	var request blockRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.social.BlockPlayer(request.PlayerID, request.BlockedPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handleListBlocks(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	blocks, appErr := h.social.ListBlocks(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"player_id": playerID,
		"blocks":    blocks,
	})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
