package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/chat/internal/domain"
)

const (
	kindPrivate = "private"
	kindGroup   = "group"
	kindGuild   = "guild"
	kindParty   = "party"
	kindWorld   = "world"
	kindSystem  = "system"
	kindCustom  = "custom"

	defaultReplayLimit = 50
	maxReplayLimit     = 200
	deliveryModePush   = "online_push"
	deliveryModeReplay = "offline_replay"
	presenceOnline     = "online"
	presenceOffline    = "offline"
)

// PresenceSnapshot contains the subset of presence state chat uses for planning.
type PresenceSnapshot struct {
	PlayerID  string `json:"player_id"`
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// PresenceReader resolves current presence state for delivery planning.
type PresenceReader interface {
	GetPresence(ctx context.Context, playerID string) (PresenceSnapshot, *apperrors.Error)
}

// DeliveryTarget describes how chat would route a message to a member.
type DeliveryTarget struct {
	PlayerID     string `json:"player_id"`
	Presence     string `json:"presence"`
	DeliveryMode string `json:"delivery_mode"`
	SessionID    string `json:"session_id,omitempty"`
	RealmID      string `json:"realm_id,omitempty"`
	Location     string `json:"location,omitempty"`
}

// ChatService provides an in-memory prototype for conversation, seq, ack, and replay flows.
type ChatService struct {
	mu                sync.RWMutex
	conversations     map[string]domain.Conversation
	messages          map[string][]domain.Message
	readCursors       map[string]map[string]domain.ReadCursor
	presence          PresenceReader
	now               func() time.Time
	newConversationID func() (string, error)
	newMessageID      func() (string, error)
}

// NewChatService constructs an in-memory chat service.
func NewChatService(presence PresenceReader) *ChatService {
	return &ChatService{
		conversations: make(map[string]domain.Conversation),
		messages:      make(map[string][]domain.Message),
		readCursors:   make(map[string]map[string]domain.ReadCursor),
		presence:      presence,
		now:           time.Now,
		newConversationID: func() (string, error) {
			return idgen.Token(8)
		},
		newMessageID: func() (string, error) {
			return idgen.Token(10)
		},
	}
}

// CreateConversation creates a conversation with explicit member scope.
func (s *ChatService) CreateConversation(kind string, resourceID string, memberPlayerIDs []string) (domain.Conversation, *apperrors.Error) {
	normalizedMembers, err := normalizeMembers(kind, memberPlayerIDs)
	if err != nil {
		return domain.Conversation{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, conversation := range s.conversations {
		if conversation.Kind == kind &&
			conversation.ResourceID == resourceID &&
			slices.Equal(conversation.MemberPlayerIDs, normalizedMembers) {
			return conversation, nil
		}
	}

	conversationID, idErr := s.newConversationID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.Conversation{}, &internal
	}

	conversation := domain.Conversation{
		ID:              conversationID,
		Kind:            kind,
		ResourceID:      resourceID,
		MemberPlayerIDs: normalizedMembers,
		LastSeq:         0,
		CreatedAt:       s.now(),
	}

	s.conversations[conversation.ID] = conversation
	return conversation, nil
}

// ListConversations returns conversations visible to the given player.
func (s *ChatService) ListConversations(playerID string) ([]domain.Conversation, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	conversations := make([]domain.Conversation, 0)
	for _, conversation := range s.conversations {
		if hasMember(conversation.MemberPlayerIDs, playerID) || conversation.Kind == kindSystem {
			conversations = append(conversations, conversation)
		}
	}

	slices.SortFunc(conversations, func(a domain.Conversation, b domain.Conversation) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}

		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})

	return conversations, nil
}

// SendMessage appends a new message and advances conversation seq.
func (s *ChatService) SendMessage(conversationID string, senderPlayerID string, body string) (domain.Message, *apperrors.Error) {
	if conversationID == "" || senderPlayerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and sender_player_id are required", http.StatusBadRequest)
		return domain.Message{}, &err
	}

	body = strings.TrimSpace(body)
	if body == "" {
		err := apperrors.New("invalid_request", "body is required", http.StatusBadRequest)
		return domain.Message{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	conversation, ok := s.conversations[conversationID]
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return domain.Message{}, &err
	}

	if appErr := validateSendPermission(conversation, senderPlayerID); appErr != nil {
		return domain.Message{}, appErr
	}

	messageID, idErr := s.newMessageID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.Message{}, &internal
	}

	conversation.LastSeq++
	s.conversations[conversation.ID] = conversation

	message := domain.Message{
		ID:             messageID,
		ConversationID: conversation.ID,
		Seq:            conversation.LastSeq,
		SenderPlayerID: senderPlayerID,
		Body:           body,
		CreatedAt:      s.now(),
	}

	s.messages[conversation.ID] = append(s.messages[conversation.ID], message)
	return message, nil
}

// AckConversation advances a player's read cursor.
func (s *ChatService) AckConversation(conversationID string, playerID string, ackSeq int64) (domain.ReadCursor, *apperrors.Error) {
	if conversationID == "" || playerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and player_id are required", http.StatusBadRequest)
		return domain.ReadCursor{}, &err
	}

	if ackSeq < 0 {
		err := apperrors.New("invalid_request", "ack_seq must be >= 0", http.StatusBadRequest)
		return domain.ReadCursor{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	conversation, ok := s.conversations[conversationID]
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return domain.ReadCursor{}, &err
	}

	if !hasMember(conversation.MemberPlayerIDs, playerID) && conversation.Kind != kindSystem {
		err := apperrors.New("forbidden", "player is not a member of the conversation", http.StatusForbidden)
		return domain.ReadCursor{}, &err
	}

	if ackSeq > conversation.LastSeq {
		err := apperrors.New("invalid_request", "ack_seq cannot exceed conversation last_seq", http.StatusBadRequest)
		return domain.ReadCursor{}, &err
	}

	if s.readCursors[conversationID] == nil {
		s.readCursors[conversationID] = make(map[string]domain.ReadCursor)
	}

	cursor := s.readCursors[conversationID][playerID]
	if ackSeq < cursor.AckSeq {
		ackSeq = cursor.AckSeq
	}

	cursor = domain.ReadCursor{
		ConversationID: conversationID,
		PlayerID:       playerID,
		AckSeq:         ackSeq,
		UpdatedAt:      s.now(),
	}

	s.readCursors[conversationID][playerID] = cursor
	return cursor, nil
}

// ReplayMessages returns messages with seq greater than afterSeq.
func (s *ChatService) ReplayMessages(conversationID string, playerID string, afterSeq int64, limit int) ([]domain.Message, *apperrors.Error) {
	if conversationID == "" || playerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and player_id are required", http.StatusBadRequest)
		return nil, &err
	}

	if afterSeq < 0 {
		err := apperrors.New("invalid_request", "after_seq must be >= 0", http.StatusBadRequest)
		return nil, &err
	}

	if limit <= 0 {
		limit = defaultReplayLimit
	}
	if limit > maxReplayLimit {
		limit = maxReplayLimit
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	conversation, ok := s.conversations[conversationID]
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return nil, &err
	}

	if !hasMember(conversation.MemberPlayerIDs, playerID) && conversation.Kind != kindSystem {
		err := apperrors.New("forbidden", "player is not a member of the conversation", http.StatusForbidden)
		return nil, &err
	}

	messages := s.messages[conversationID]
	replay := make([]domain.Message, 0, min(limit, len(messages)))
	for _, message := range messages {
		if message.Seq <= afterSeq {
			continue
		}

		replay = append(replay, message)
		if len(replay) == limit {
			break
		}
	}

	return replay, nil
}

// PlanDelivery resolves current routing mode for other members in the conversation.
func (s *ChatService) PlanDelivery(ctx context.Context, conversationID string, senderPlayerID string) ([]DeliveryTarget, *apperrors.Error) {
	if conversationID == "" || senderPlayerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and sender_player_id are required", http.StatusBadRequest)
		return nil, &err
	}

	s.mu.RLock()
	conversation, ok := s.conversations[conversationID]
	s.mu.RUnlock()
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return nil, &err
	}

	if appErr := validateSendPermission(conversation, senderPlayerID); appErr != nil {
		return nil, appErr
	}

	targets := make([]DeliveryTarget, 0, len(conversation.MemberPlayerIDs))
	for _, memberID := range conversation.MemberPlayerIDs {
		if memberID == senderPlayerID {
			continue
		}

		target := DeliveryTarget{
			PlayerID:     memberID,
			Presence:     presenceOffline,
			DeliveryMode: deliveryModeReplay,
		}

		if s.presence == nil {
			targets = append(targets, target)
			continue
		}

		snapshot, appErr := s.presence.GetPresence(ctx, memberID)
		if appErr != nil {
			if appErr.Code == "not_found" {
				targets = append(targets, target)
				continue
			}
			return nil, appErr
		}

		target.Presence = snapshot.Status
		target.SessionID = snapshot.SessionID
		target.RealmID = snapshot.RealmID
		target.Location = snapshot.Location
		if snapshot.Status == presenceOnline {
			target.DeliveryMode = deliveryModePush
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func normalizeMembers(kind string, members []string) ([]string, *apperrors.Error) {
	if !isSupportedKind(kind) {
		err := apperrors.New("invalid_request", "unsupported conversation kind", http.StatusBadRequest)
		return nil, &err
	}

	seen := make(map[string]struct{}, len(members))
	normalized := make([]string, 0, len(members))
	for _, memberID := range members {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			continue
		}
		if _, ok := seen[memberID]; ok {
			continue
		}
		seen[memberID] = struct{}{}
		normalized = append(normalized, memberID)
	}

	slices.Sort(normalized)

	switch kind {
	case kindPrivate:
		if len(normalized) != 2 {
			err := apperrors.New("invalid_request", "private conversations require exactly 2 members", http.StatusBadRequest)
			return nil, &err
		}
	case kindSystem:
		if len(normalized) == 0 {
			err := apperrors.New("invalid_request", "system conversations require at least 1 member", http.StatusBadRequest)
			return nil, &err
		}
	default:
		if len(normalized) == 0 {
			err := apperrors.New("invalid_request", "conversation requires at least 1 member", http.StatusBadRequest)
			return nil, &err
		}
	}

	return normalized, nil
}

func validateSendPermission(conversation domain.Conversation, senderPlayerID string) *apperrors.Error {
	switch conversation.Kind {
	case kindSystem:
		if senderPlayerID != "system" {
			err := apperrors.New("forbidden", "system conversations only accept system sender", http.StatusForbidden)
			return &err
		}
		return nil
	case kindWorld, kindGuild, kindParty, kindPrivate, kindGroup, kindCustom:
		if !hasMember(conversation.MemberPlayerIDs, senderPlayerID) {
			err := apperrors.New("forbidden", "sender is not allowed in the conversation", http.StatusForbidden)
			return &err
		}
		return nil
	default:
		err := apperrors.New("invalid_request", "unsupported conversation kind", http.StatusBadRequest)
		return &err
	}
}

func hasMember(members []string, playerID string) bool {
	return slices.Contains(members, playerID)
}

func isSupportedKind(kind string) bool {
	switch kind {
	case kindPrivate, kindGroup, kindGuild, kindParty, kindWorld, kindSystem, kindCustom:
		return true
	default:
		return false
	}
}

func (s *ChatService) String() string {
	return fmt.Sprintf("chat-service(conversations=%d)", len(s.conversations))
}
