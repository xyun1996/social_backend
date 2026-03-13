package service

import (
	"context"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// ChatAcker exposes the chat ack boundary for gateway.
type ChatAcker interface {
	AckConversation(ctx context.Context, conversationID string, playerID string, ackSeq int64) *apperrors.Error
}

// AckRequest is the gateway-owned realtime ack request.
type AckRequest struct {
	SessionID      string `json:"session_id"`
	ConversationID string `json:"conversation_id"`
	AckSeq         int64  `json:"ack_seq"`
}

// AckResult summarizes both chat cursor advancement and gateway-local inbox pruning.
type AckResult struct {
	SessionID        string `json:"session_id"`
	ConversationID   string `json:"conversation_id"`
	AckSeq           int64  `json:"ack_seq"`
	PrunedCount      int    `json:"pruned_count"`
	LastAckedEventID string `json:"last_acked_event_id,omitempty"`
}

// AckService validates session ownership before forwarding chat acks.
type AckService struct {
	realtime *RealtimeService
	chat     ChatAcker
}

// NewAckService constructs the gateway chat ack service.
func NewAckService(realtime *RealtimeService, chat ChatAcker) *AckService {
	return &AckService{realtime: realtime, chat: chat}
}

// AckConversation forwards a session-scoped ack to chat and compacts gateway-local inbox state.
func (s *AckService) AckConversation(ctx context.Context, request AckRequest) (AckResult, *apperrors.Error) {
	if request.SessionID == "" || request.ConversationID == "" {
		err := apperrors.New("invalid_request", "session_id and conversation_id are required", http.StatusBadRequest)
		return AckResult{}, &err
	}
	if request.AckSeq < 0 {
		err := apperrors.New("invalid_request", "ack_seq must be >= 0", http.StatusBadRequest)
		return AckResult{}, &err
	}
	if s.realtime == nil || s.chat == nil {
		err := apperrors.New("dependency_missing", "gateway ack dependencies are not configured", http.StatusInternalServerError)
		return AckResult{}, &err
	}

	session, appErr := s.realtime.GetSession(request.SessionID)
	if appErr != nil {
		return AckResult{}, appErr
	}
	if session.State != sessionStateActive {
		err := apperrors.New("invalid_state", "session is not active", http.StatusConflict)
		return AckResult{}, &err
	}

	if appErr := s.chat.AckConversation(ctx, request.ConversationID, session.PlayerID, request.AckSeq); appErr != nil {
		return AckResult{}, appErr
	}

	compaction, appErr := s.realtime.AcknowledgeConversation(request.SessionID, request.ConversationID, request.AckSeq)
	if appErr != nil {
		return AckResult{}, appErr
	}

	return AckResult{
		SessionID:        compaction.SessionID,
		ConversationID:   compaction.ConversationID,
		AckSeq:           compaction.AckSeq,
		PrunedCount:      compaction.PrunedCount,
		LastAckedEventID: compaction.LastAckedEventID,
	}, nil
}
