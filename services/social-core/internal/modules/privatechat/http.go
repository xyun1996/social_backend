package privatechat

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	mux.HandleFunc("POST /v1/private-chat/conversations", h.handleCreateConversation)
	mux.HandleFunc("GET /v1/private-chat/conversations", h.handleListConversations)
	mux.HandleFunc("GET /v1/private-chat/summaries", h.handleListSummaries)
	mux.HandleFunc("GET /v1/private-chat/conversations/{conversationID}/summary", h.handleGetSummary)
	mux.HandleFunc("POST /v1/private-chat/conversations/{conversationID}/messages", h.handleSendMessage)
	mux.HandleFunc("GET /v1/private-chat/conversations/{conversationID}/messages", h.handleReplayMessages)
	mux.HandleFunc("POST /v1/private-chat/conversations/{conversationID}/ack", h.handleAckConversation)
}

func (h *HTTPHandler) handleCreateConversation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		MemberPlayerIDs []string `json:"member_player_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.CreateConversation(request.MemberPlayerIDs)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleListConversations(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	records, appErr := h.service.ListConversations(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"player_id": playerID, "count": len(records), "conversations": records})
}

func (h *HTTPHandler) handleListSummaries(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	records, appErr := h.service.ListSummaries(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{"player_id": playerID, "count": len(records), "summaries": records})
}

func (h *HTTPHandler) handleGetSummary(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	record, appErr := h.service.GetSummary(r.PathValue("conversationID"), playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SenderPlayerID string `json:"sender_player_id"`
		Body           string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	record, appErr := h.service.SendMessage(r.PathValue("conversationID"), request.SenderPlayerID, request.Body)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, record)
}

func (h *HTTPHandler) handleReplayMessages(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	afterSeq, err := parseInt64Query(r, "after_seq")
	if err != nil {
		transport.WriteError(w, invalidQueryError("after_seq must be an integer"))
		return
	}
	limit, err := parseIntQuery(r, "limit")
	if err != nil {
		transport.WriteError(w, invalidQueryError("limit must be an integer"))
		return
	}
	records, appErr := h.service.ReplayMessages(r.PathValue("conversationID"), playerID, afterSeq, limit)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"conversation_id": r.PathValue("conversationID"),
		"player_id":       playerID,
		"after_seq":       afterSeq,
		"count":           len(records),
		"messages":        records,
	})
}

func (h *HTTPHandler) handleAckConversation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID string `json:"player_id"`
		AckSeq   int64  `json:"ack_seq"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}
	if appErr := h.service.AckConversation(r.PathValue("conversationID"), request.PlayerID, request.AckSeq); appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}
	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"conversation_id": r.PathValue("conversationID"),
		"player_id":       request.PlayerID,
		"ack_seq":         request.AckSeq,
	})
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}

func invalidQueryError(message string) apperrors.Error {
	return apperrors.New("invalid_query", message, http.StatusBadRequest)
}

func parseInt64Query(r *http.Request, key string) (int64, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseIntQuery(r *http.Request, key string) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}
