package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/chat/internal/service"
)

// HTTPHandler exposes the early chat HTTP API.
type HTTPHandler struct {
	chat *service.ChatService
}

// NewHTTPHandler constructs the chat HTTP routes.
func NewHTTPHandler(chat *service.ChatService) *HTTPHandler {
	return &HTTPHandler{chat: chat}
}

type createConversationRequest struct {
	Kind            string   `json:"kind"`
	ResourceID      string   `json:"resource_id"`
	MemberPlayerIDs []string `json:"member_player_ids"`
}

type sendMessageRequest struct {
	SenderPlayerID string `json:"sender_player_id"`
	Body           string `json:"body"`
}

type ackRequest struct {
	PlayerID string `json:"player_id"`
	AckSeq   int64  `json:"ack_seq"`
}

type offlineDeliveryRequest struct {
	ConversationID  string `json:"conversation_id"`
	MessageID       string `json:"message_id"`
	RecipientPlayer string `json:"recipient_player"`
	DeliveryMode    string `json:"delivery_mode"`
}

// Routes returns the chat HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/conversations", h.handleCreateConversation)
	mux.HandleFunc("GET /v1/conversations", h.handleListConversations)
	mux.HandleFunc("POST /v1/conversations/{conversationID}/messages", h.handleSendMessage)
	mux.HandleFunc("GET /v1/conversations/{conversationID}/messages", h.handleReplayMessages)
	mux.HandleFunc("POST /v1/conversations/{conversationID}/ack", h.handleAckConversation)
	mux.HandleFunc("GET /v1/conversations/{conversationID}/delivery", h.handleDeliveryPlan)
	mux.HandleFunc("POST /v1/internal/offline-deliveries", h.handleRecordOfflineDelivery)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "chat",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleCreateConversation(w http.ResponseWriter, r *http.Request) {
	var request createConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	conversation, appErr := h.chat.CreateConversation(request.Kind, request.ResourceID, request.MemberPlayerIDs)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, conversation)
}

func (h *HTTPHandler) handleListConversations(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	conversations, appErr := h.chat.ListConversations(playerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"player_id":     playerID,
		"count":         len(conversations),
		"conversations": conversations,
	})
}

func (h *HTTPHandler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var request sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	message, appErr := h.chat.SendMessage(r.PathValue("conversationID"), request.SenderPlayerID, request.Body)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, message)
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

	messages, appErr := h.chat.ReplayMessages(r.PathValue("conversationID"), playerID, afterSeq, limit)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"conversation_id": r.PathValue("conversationID"),
		"player_id":       playerID,
		"after_seq":       afterSeq,
		"count":           len(messages),
		"messages":        messages,
	})
}

func (h *HTTPHandler) handleAckConversation(w http.ResponseWriter, r *http.Request) {
	var request ackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	cursor, appErr := h.chat.AckConversation(r.PathValue("conversationID"), request.PlayerID, request.AckSeq)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, cursor)
}

func (h *HTTPHandler) handleDeliveryPlan(w http.ResponseWriter, r *http.Request) {
	senderPlayerID := r.URL.Query().Get("sender_player_id")
	targets, appErr := h.chat.PlanDelivery(r.Context(), r.PathValue("conversationID"), senderPlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"conversation_id":  r.PathValue("conversationID"),
		"sender_player_id": senderPlayerID,
		"count":            len(targets),
		"targets":          targets,
	})
}

func (h *HTTPHandler) handleRecordOfflineDelivery(w http.ResponseWriter, r *http.Request) {
	var request offlineDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	receipt, appErr := h.chat.RecordOfflineDelivery(map[string]any{
		"conversation_id":  request.ConversationID,
		"message_id":       request.MessageID,
		"recipient_player": request.RecipientPlayer,
		"delivery_mode":    request.DeliveryMode,
	})
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, receipt)
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
