package service

import (
	"context"
	"fmt"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

const (
	deliveryModePush   = "online_push"
	deliveryModeReplay = "offline_replay"
)

// ChatDeliveryTarget is the subset of chat delivery planning used by gateway.
type ChatDeliveryTarget struct {
	PlayerID     string `json:"player_id"`
	Presence     string `json:"presence"`
	DeliveryMode string `json:"delivery_mode"`
	SessionID    string `json:"session_id,omitempty"`
	RealmID      string `json:"realm_id,omitempty"`
	Location     string `json:"location,omitempty"`
}

// ChatPlanner exposes the chat delivery planning boundary for gateway.
type ChatPlanner interface {
	PlanDelivery(ctx context.Context, conversationID string, senderPlayerID string) ([]ChatDeliveryTarget, *apperrors.Error)
}

// ChatMessageEnvelope is the gateway-owned realtime event shape for chat delivery.
type ChatMessageEnvelope struct {
	EventID        string `json:"event_id"`
	Stream         string `json:"stream"`
	Kind           string `json:"kind"`
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Seq            int64  `json:"seq"`
	SenderPlayerID string `json:"sender_player_id"`
	Body           string `json:"body"`
	SentAt         string `json:"sent_at"`
}

// ChatDispatchRequest describes a post-chat-send delivery dispatch request.
type ChatDispatchRequest struct {
	ConversationID string `json:"conversation_id"`
	SenderPlayerID string `json:"sender_player_id"`
	MessageID      string `json:"message_id"`
	Seq            int64  `json:"seq"`
	Body           string `json:"body"`
	SentAt         string `json:"sent_at"`
}

// ChatDispatchResult summarizes gateway-side direct versus deferred routing.
type ChatDispatchResult struct {
	ConversationID string               `json:"conversation_id"`
	MessageID      string               `json:"message_id"`
	PushedCount    int                  `json:"pushed_count"`
	DeferredCount  int                  `json:"deferred_count"`
	PushedSessions []string             `json:"pushed_sessions"`
	Deferred       []ChatDeliveryTarget `json:"deferred"`
	Targets        []ChatDeliveryTarget `json:"targets"`
}

// DeliveryService routes chat delivery plans into gateway session inboxes.
type DeliveryService struct {
	realtime *RealtimeService
	planner  ChatPlanner
}

// NewDeliveryService constructs the gateway delivery prototype.
func NewDeliveryService(realtime *RealtimeService, planner ChatPlanner) *DeliveryService {
	return &DeliveryService{realtime: realtime, planner: planner}
}

// DispatchChat routes online chat recipients into session inboxes and defers offline ones.
func (s *DeliveryService) DispatchChat(ctx context.Context, request ChatDispatchRequest) (ChatDispatchResult, *apperrors.Error) {
	if request.ConversationID == "" || request.SenderPlayerID == "" || request.MessageID == "" {
		err := apperrors.New("invalid_request", "conversation_id, sender_player_id, and message_id are required", http.StatusBadRequest)
		return ChatDispatchResult{}, &err
	}
	if s.realtime == nil || s.planner == nil {
		err := apperrors.New("dependency_missing", "gateway delivery dependencies are not configured", http.StatusInternalServerError)
		return ChatDispatchResult{}, &err
	}

	targets, appErr := s.planner.PlanDelivery(ctx, request.ConversationID, request.SenderPlayerID)
	if appErr != nil {
		return ChatDispatchResult{}, appErr
	}

	result := ChatDispatchResult{
		ConversationID: request.ConversationID,
		MessageID:      request.MessageID,
		Targets:        targets,
		PushedSessions: make([]string, 0),
		Deferred:       make([]ChatDeliveryTarget, 0),
	}

	envelope := ChatMessageEnvelope{
		EventID:        fmt.Sprintf("%s:%d", request.MessageID, request.Seq),
		Stream:         "chat",
		Kind:           "chat.message",
		ConversationID: request.ConversationID,
		MessageID:      request.MessageID,
		Seq:            request.Seq,
		SenderPlayerID: request.SenderPlayerID,
		Body:           request.Body,
		SentAt:         request.SentAt,
	}

	for _, target := range targets {
		if target.DeliveryMode != deliveryModePush || target.SessionID == "" {
			result.Deferred = append(result.Deferred, target)
			continue
		}

		if appErr := s.realtime.EnqueueChatEvent(target.SessionID, envelope); appErr != nil {
			target.DeliveryMode = deliveryModeReplay
			result.Deferred = append(result.Deferred, target)
			continue
		}

		result.PushedSessions = append(result.PushedSessions, target.SessionID)
	}

	result.PushedCount = len(result.PushedSessions)
	result.DeferredCount = len(result.Deferred)
	return result, nil
}
