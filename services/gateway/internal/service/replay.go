package service

import (
	"context"
	"net/http"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// ChatReplayMessage is the subset of chat replay data gateway exposes to sessions.
type ChatReplayMessage struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Seq            int64     `json:"seq"`
	SenderPlayerID string    `json:"sender_player_id"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

// ChatReplayer exposes the chat replay boundary for gateway.
type ChatReplayer interface {
	ReplayMessages(ctx context.Context, conversationID string, playerID string, afterSeq int64, limit int) ([]ChatReplayMessage, *apperrors.Error)
}

// ReplayRequest is the session-scoped replay handoff request.
type ReplayRequest struct {
	SessionID      string `json:"session_id"`
	ConversationID string `json:"conversation_id"`
	AfterSeq       int64  `json:"after_seq"`
	Limit          int    `json:"limit"`
}

// ReplayResult returns replay data using the active session identity.
type ReplayResult struct {
	SessionID      string              `json:"session_id"`
	ConversationID string              `json:"conversation_id"`
	PlayerID       string              `json:"player_id"`
	AfterSeq       int64               `json:"after_seq"`
	Count          int                 `json:"count"`
	Messages       []ChatReplayMessage `json:"messages"`
}

// ReplayService resolves replay through chat using the active session subject.
type ReplayService struct {
	realtime *RealtimeService
	chat     ChatReplayer
}

// NewReplayService constructs the gateway replay service.
func NewReplayService(realtime *RealtimeService, chat ChatReplayer) *ReplayService {
	return &ReplayService{realtime: realtime, chat: chat}
}

// ReplayConversation returns messages after the requested sequence for the active session player.
func (s *ReplayService) ReplayConversation(ctx context.Context, request ReplayRequest) (ReplayResult, *apperrors.Error) {
	if request.SessionID == "" || request.ConversationID == "" {
		err := apperrors.New("invalid_request", "session_id and conversation_id are required", http.StatusBadRequest)
		return ReplayResult{}, &err
	}
	if request.AfterSeq < 0 {
		err := apperrors.New("invalid_request", "after_seq must be >= 0", http.StatusBadRequest)
		return ReplayResult{}, &err
	}
	if s.realtime == nil || s.chat == nil {
		err := apperrors.New("dependency_missing", "gateway replay dependencies are not configured", http.StatusInternalServerError)
		return ReplayResult{}, &err
	}

	session, appErr := s.realtime.getActiveSession(request.SessionID)
	if appErr != nil {
		return ReplayResult{}, appErr
	}

	messages, appErr := s.chat.ReplayMessages(ctx, request.ConversationID, session.PlayerID, request.AfterSeq, request.Limit)
	if appErr != nil {
		return ReplayResult{}, appErr
	}

	return ReplayResult{
		SessionID:      request.SessionID,
		ConversationID: request.ConversationID,
		PlayerID:       session.PlayerID,
		AfterSeq:       request.AfterSeq,
		Count:          len(messages),
		Messages:       messages,
	}, nil
}
