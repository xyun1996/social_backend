package service

import (
	"context"
	"encoding/json"
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

	defaultReplayLimit   = 50
	maxReplayLimit       = 200
	deliveryModePush     = "online_push"
	deliveryModeReplay   = "offline_replay"
	presenceOnline       = "online"
	presenceOffline      = "offline"
	offlineJobType       = "chat.offline_delivery"
	channelScopeDirect   = "direct"
	channelScopeResource = "resource"
	membershipExplicit   = "explicit"
	membershipBound      = "resource_bound"
	sendPolicyMembers    = "members"
	sendPolicyModerated  = "moderated"
	sendPolicySystemOnly = "system_only"
	visibilityMembers   = "members"
	visibilityPublicRead = "public_read"
	moderationOpen      = "open"
	moderationManaged   = "managed"
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

// JobScheduler captures async scheduling intent for chat follow-up work.
type JobScheduler interface {
	EnqueueJob(ctx context.Context, jobType string, payload string) *apperrors.Error
}

// GuildMembershipReader resolves whether a player currently belongs to a guild.
type GuildMembershipReader interface {
	IsGuildMember(ctx context.Context, guildID string, playerID string) (bool, *apperrors.Error)
}

// PartyMembershipReader resolves whether a player currently belongs to a party.
type PartyMembershipReader interface {
	IsPartyMember(ctx context.Context, partyID string, playerID string) (bool, *apperrors.Error)
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

// OfflineDeliveryReceipt records worker-side offline delivery processing.
type OfflineDeliveryReceipt struct {
	ConversationID  string `json:"conversation_id"`
	MessageID       string `json:"message_id"`
	RecipientPlayer string `json:"recipient_player"`
	DeliveryMode    string `json:"delivery_mode"`
	ProcessedAt     string `json:"processed_at"`
}

// ChatService provides an in-memory prototype for conversation, seq, ack, and replay flows.
type ChatService struct {
	offlineMu         sync.RWMutex
	offlineDeliveries []OfflineDeliveryReceipt
	conversations     ConversationStore
	messages          MessageStore
	readCursors       ReadCursorStore
	presence          PresenceReader
	scheduler         JobScheduler
	guilds            GuildMembershipReader
	parties           PartyMembershipReader
	now               func() time.Time
	newConversationID func() (string, error)
	newMessageID      func() (string, error)
}

// NewChatService constructs an in-memory chat service.
func NewChatService(presence PresenceReader, scheduler JobScheduler) *ChatService {
	return &ChatService{
		conversations:     newMemoryConversationStore(),
		messages:          newMemoryMessageStore(),
		readCursors:       newMemoryReadCursorStore(),
		offlineDeliveries: make([]OfflineDeliveryReceipt, 0),
		presence:          presence,
		scheduler:         scheduler,
		now:               time.Now,
		newConversationID: func() (string, error) {
			return idgen.Token(8)
		},
		newMessageID: func() (string, error) {
			return idgen.Token(10)
		},
	}
}

// NewChatServiceWithStores constructs a chat service with custom durable stores.
func NewChatServiceWithStores(conversations ConversationStore, messages MessageStore, readCursors ReadCursorStore, presence PresenceReader, scheduler JobScheduler) *ChatService {
	if conversations == nil || messages == nil || readCursors == nil {
		return NewChatService(presence, scheduler)
	}

	return &ChatService{
		conversations:     conversations,
		messages:          messages,
		readCursors:       readCursors,
		offlineDeliveries: make([]OfflineDeliveryReceipt, 0),
		presence:          presence,
		scheduler:         scheduler,
		guilds:            nil,
		parties:           nil,
		now:               time.Now,
		newConversationID: func() (string, error) { return idgen.Token(8) },
		newMessageID:      func() (string, error) { return idgen.Token(10) },
	}
}

// SetMembershipReaders wires optional resource membership readers for guild and party channels.
func (s *ChatService) SetMembershipReaders(guilds GuildMembershipReader, parties PartyMembershipReader) {
	if s == nil {
		return
	}
	s.guilds = guilds
	s.parties = parties
}

// CreateConversation creates a conversation with explicit member scope.
func (s *ChatService) CreateConversation(kind string, resourceID string, memberPlayerIDs []string) (domain.Conversation, *apperrors.Error) {
	normalizedMembers, normalizedResourceID, err := normalizeMembers(kind, resourceID, memberPlayerIDs)
	if err != nil {
		return domain.Conversation{}, err
	}

	conversations, storeErr := s.conversations.ListConversations()
	if storeErr != nil {
		internal := apperrors.Internal()
		return domain.Conversation{}, &internal
	}

	for _, conversation := range conversations {
		if conversation.Kind != kind {
			continue
		}

		if isResourceBoundKind(kind) {
			if conversation.ResourceID != normalizedResourceID {
				continue
			}

			mergedMembers := mergeMembers(conversation.MemberPlayerIDs, normalizedMembers)
			if !slices.Equal(conversation.MemberPlayerIDs, mergedMembers) {
				conversation.MemberPlayerIDs = mergedMembers
				if err := s.conversations.SaveConversation(conversation); err != nil {
					internal := apperrors.Internal()
					return domain.Conversation{}, &internal
				}
			}
			return conversation, nil
		}

		if conversation.ResourceID == normalizedResourceID &&
			slices.Equal(conversation.MemberPlayerIDs, normalizedMembers) {
			return conversation, nil
		}
	}

	conversationID, idErr := s.newConversationID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.Conversation{}, &internal
	}

	conversation := defaultConversationGovernance(domain.Conversation{
		ID:              conversationID,
		Kind:            kind,
		ResourceID:      normalizedResourceID,
		MemberPlayerIDs: normalizedMembers,
		LastSeq:         0,
		CreatedAt:       s.now(),
	})

	if err := s.conversations.SaveConversation(conversation); err != nil {
		internal := apperrors.Internal()
		return domain.Conversation{}, &internal
	}
	return conversation, nil
}

// GetChannelDescriptor returns the resolved policy surface for a stored conversation.
func (s *ChatService) GetChannelDescriptor(conversationID string) (domain.ChannelDescriptor, *apperrors.Error) {
	if conversationID == "" {
		err := apperrors.New("invalid_request", "conversation_id is required", http.StatusBadRequest)
		return domain.ChannelDescriptor{}, &err
	}

	conversation, ok, err := s.conversations.GetConversation(conversationID)
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return domain.ChannelDescriptor{}, &err
	}
	if err != nil {
		internal := apperrors.Internal()
		return domain.ChannelDescriptor{}, &internal
	}

	return buildChannelDescriptor(conversation), nil
}

// GetConversationSummary returns unread and last-message state for one player-conversation pair.
func (s *ChatService) GetConversationSummary(conversationID string, playerID string) (domain.ConversationSummary, *apperrors.Error) {
	if conversationID == "" || playerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and player_id are required", http.StatusBadRequest)
		return domain.ConversationSummary{}, &err
	}

	conversation, ok, err := s.conversations.GetConversation(conversationID)
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return domain.ConversationSummary{}, &err
	}
	if err != nil {
		internal := apperrors.Internal()
		return domain.ConversationSummary{}, &internal
	}
	allowed, appErr := s.canAccessConversation(context.Background(), conversation, playerID)
	if appErr != nil {
		return domain.ConversationSummary{}, appErr
	}
	if !allowed {
		err := apperrors.New("forbidden", "player is not allowed in the conversation", http.StatusForbidden)
		return domain.ConversationSummary{}, &err
	}

	return s.buildConversationSummary(conversation, playerID)
}

// ListConversationSummaries returns all visible conversation summaries for a player.
func (s *ChatService) ListConversationSummaries(playerID string) ([]domain.ConversationSummary, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}

	conversations, appErr := s.ListConversations(playerID)
	if appErr != nil {
		return nil, appErr
	}

	summaries := make([]domain.ConversationSummary, 0, len(conversations))
	for _, conversation := range conversations {
		summary, appErr := s.buildConversationSummary(conversation, playerID)
		if appErr != nil {
			return nil, appErr
		}
		summaries = append(summaries, summary)
	}
	slices.SortFunc(summaries, func(a domain.ConversationSummary, b domain.ConversationSummary) int {
		if !a.UpdatedAt.Equal(b.UpdatedAt) {
			if a.UpdatedAt.After(b.UpdatedAt) {
				return -1
			}
			return 1
		}
		switch {
		case a.ConversationID < b.ConversationID:
			return -1
		case a.ConversationID > b.ConversationID:
			return 1
		default:
			return 0
		}
	})
	return summaries, nil
}

// ListConversations returns conversations visible to the given player.
func (s *ChatService) ListConversations(playerID string) ([]domain.Conversation, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}

	allConversations, storeErr := s.conversations.ListConversations()
	if storeErr != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	conversations := make([]domain.Conversation, 0)
	for _, conversation := range allConversations {
		allowed, appErr := s.canAccessConversation(context.Background(), conversation, playerID)
		if appErr != nil {
			return nil, appErr
		}
		if allowed {
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

	conversation, ok, err := s.conversations.GetConversation(conversationID)
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return domain.Message{}, &err
	}
	if err != nil {
		internal := apperrors.Internal()
		return domain.Message{}, &internal
	}

	if appErr := s.validateSendPermission(context.Background(), conversation, senderPlayerID); appErr != nil {
		return domain.Message{}, appErr
	}

	messageID, idErr := s.newMessageID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.Message{}, &internal
	}

	conversation.LastSeq++
	if err := s.conversations.SaveConversation(conversation); err != nil {
		internal := apperrors.Internal()
		return domain.Message{}, &internal
	}

	message := domain.Message{
		ID:             messageID,
		ConversationID: conversation.ID,
		Seq:            conversation.LastSeq,
		SenderPlayerID: senderPlayerID,
		Body:           body,
		CreatedAt:      s.now(),
	}

	if err := s.messages.AppendMessage(message); err != nil {
		internal := apperrors.Internal()
		return domain.Message{}, &internal
	}

	if s.scheduler != nil {
		targets, appErr := s.PlanDelivery(context.Background(), conversation.ID, senderPlayerID)
		if appErr == nil {
			s.enqueueOfflineDeliveries(context.Background(), conversation, message, targets)
		}
	}

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

	conversation, ok, err := s.conversations.GetConversation(conversationID)
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return domain.ReadCursor{}, &err
	}
	if err != nil {
		internal := apperrors.Internal()
		return domain.ReadCursor{}, &internal
	}

	allowed, appErr := s.canAccessConversation(context.Background(), conversation, playerID)
	if appErr != nil {
		return domain.ReadCursor{}, appErr
	}
	if !allowed {
		err := apperrors.New("forbidden", "player is not allowed in the conversation", http.StatusForbidden)
		return domain.ReadCursor{}, &err
	}

	if ackSeq > conversation.LastSeq {
		err := apperrors.New("invalid_request", "ack_seq cannot exceed conversation last_seq", http.StatusBadRequest)
		return domain.ReadCursor{}, &err
	}

	cursor, _, err := s.readCursors.GetCursor(conversationID, playerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.ReadCursor{}, &internal
	}
	if ackSeq < cursor.AckSeq {
		ackSeq = cursor.AckSeq
	}

	cursor = domain.ReadCursor{
		ConversationID: conversationID,
		PlayerID:       playerID,
		AckSeq:         ackSeq,
		UpdatedAt:      s.now(),
	}

	if err := s.readCursors.SaveCursor(cursor); err != nil {
		internal := apperrors.Internal()
		return domain.ReadCursor{}, &internal
	}
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

	conversation, ok, err := s.conversations.GetConversation(conversationID)
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return nil, &err
	}
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}

	allowed, appErr := s.canAccessConversation(context.Background(), conversation, playerID)
	if appErr != nil {
		return nil, appErr
	}
	if !allowed {
		err := apperrors.New("forbidden", "player is not allowed in the conversation", http.StatusForbidden)
		return nil, &err
	}

	messages, err := s.messages.ListMessages(conversationID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
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

	conversation, ok, err := s.conversations.GetConversation(conversationID)
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return nil, &err
	}
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}

	if appErr := s.validateSendPermission(ctx, conversation, senderPlayerID); appErr != nil {
		return nil, appErr
	}

	targets := make([]DeliveryTarget, 0, len(conversation.MemberPlayerIDs))
	for _, memberID := range conversation.MemberPlayerIDs {
		if memberID == senderPlayerID {
			continue
		}
		allowed, appErr := s.canAccessConversation(ctx, conversation, memberID)
		if appErr != nil {
			return nil, appErr
		}
		if !allowed {
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

// RecordOfflineDelivery stores worker-side offline delivery processing for observability.
func (s *ChatService) RecordOfflineDelivery(payload map[string]any) (OfflineDeliveryReceipt, *apperrors.Error) {
	conversationID, _ := payload["conversation_id"].(string)
	messageID, _ := payload["message_id"].(string)
	recipientPlayer, _ := payload["recipient_player"].(string)
	deliveryMode, _ := payload["delivery_mode"].(string)

	if conversationID == "" || messageID == "" || recipientPlayer == "" {
		err := apperrors.New("invalid_request", "conversation_id, message_id, and recipient_player are required", http.StatusBadRequest)
		return OfflineDeliveryReceipt{}, &err
	}

	messages, err := s.messages.ListMessages(conversationID)
	if err != nil {
		internal := apperrors.Internal()
		return OfflineDeliveryReceipt{}, &internal
	}
	found := false
	for _, message := range messages {
		if message.ID == messageID {
			found = true
			break
		}
	}
	if !found {
		err := apperrors.New("not_found", "message not found", http.StatusNotFound)
		return OfflineDeliveryReceipt{}, &err
	}

	receipt := OfflineDeliveryReceipt{
		ConversationID:  conversationID,
		MessageID:       messageID,
		RecipientPlayer: recipientPlayer,
		DeliveryMode:    deliveryMode,
		ProcessedAt:     s.now().UTC().Format(time.RFC3339Nano),
	}
	s.offlineMu.Lock()
	defer s.offlineMu.Unlock()
	s.offlineDeliveries = append(s.offlineDeliveries, receipt)
	return receipt, nil
}

// ListOfflineDeliveries returns the recorded offline delivery processing entries.
func (s *ChatService) ListOfflineDeliveries(conversationID string) []OfflineDeliveryReceipt {
	s.offlineMu.RLock()
	defer s.offlineMu.RUnlock()

	receipts := make([]OfflineDeliveryReceipt, 0)
	for _, receipt := range s.offlineDeliveries {
		if conversationID != "" && receipt.ConversationID != conversationID {
			continue
		}
		receipts = append(receipts, receipt)
	}
	return receipts
}

func normalizeMembers(kind string, resourceID string, members []string) ([]string, string, *apperrors.Error) {
	if !isSupportedKind(kind) {
		err := apperrors.New("invalid_request", "unsupported conversation kind", http.StatusBadRequest)
		return nil, "", &err
	}

	resourceID = strings.TrimSpace(resourceID)

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
		if resourceID != "" {
			err := apperrors.New("invalid_request", "private conversations cannot set resource_id", http.StatusBadRequest)
			return nil, "", &err
		}
		if len(normalized) != 2 {
			err := apperrors.New("invalid_request", "private conversations require exactly 2 members", http.StatusBadRequest)
			return nil, "", &err
		}
	case kindGroup:
		if resourceID != "" {
			err := apperrors.New("invalid_request", "group conversations cannot set resource_id", http.StatusBadRequest)
			return nil, "", &err
		}
		if len(normalized) < 2 {
			err := apperrors.New("invalid_request", "group conversations require at least 2 members", http.StatusBadRequest)
			return nil, "", &err
		}
	case kindSystem:
		if resourceID == "" {
			err := apperrors.New("invalid_request", "system conversations require resource_id", http.StatusBadRequest)
			return nil, "", &err
		}
		if len(normalized) == 0 {
			err := apperrors.New("invalid_request", "system conversations require at least 1 member", http.StatusBadRequest)
			return nil, "", &err
		}
	case kindGuild, kindParty, kindWorld, kindCustom:
		if resourceID == "" {
			err := apperrors.New("invalid_request", "resource-backed conversations require resource_id", http.StatusBadRequest)
			return nil, "", &err
		}
		if len(normalized) == 0 {
			err := apperrors.New("invalid_request", "conversation requires at least 1 member", http.StatusBadRequest)
			return nil, "", &err
		}
	}

	return normalized, resourceID, nil
}

func (s *ChatService) validateSendPermission(ctx context.Context, conversation domain.Conversation, senderPlayerID string) *apperrors.Error {
	conversation = defaultConversationGovernance(conversation)
	if senderPlayerID != "system" && isMuted(conversation, senderPlayerID) {
		err := apperrors.New("muted", "sender is muted in the conversation", http.StatusForbidden)
		return &err
	}

	switch conversation.SendPolicy {
	case sendPolicySystemOnly:
		if senderPlayerID != "system" {
			err := apperrors.New("forbidden", "system-only conversations only accept system sender", http.StatusForbidden)
			return &err
		}
		return nil
	case sendPolicyModerated:
		if senderPlayerID == "system" || isModerator(conversation, senderPlayerID) {
			return nil
		}
		err := apperrors.New("moderated", "sender must be a moderator in this conversation", http.StatusForbidden)
		return &err
	}

	switch conversation.Kind {
	case kindSystem:
		if senderPlayerID != "system" {
			err := apperrors.New("forbidden", "system conversations only accept system sender", http.StatusForbidden)
			return &err
		}
		return nil
	case kindGuild, kindParty:
		if senderPlayerID == "system" {
			return nil
		}
	}

	allowed, appErr := s.canAccessConversation(ctx, conversation, senderPlayerID)
	if appErr != nil {
		return appErr
	}
	if !allowed {
		err := apperrors.New("forbidden", "sender is not allowed in the conversation", http.StatusForbidden)
		return &err
	}
	return nil
}

func (s *ChatService) canAccessConversation(ctx context.Context, conversation domain.Conversation, playerID string) (bool, *apperrors.Error) {
	conversation = defaultConversationGovernance(conversation)
	if conversation.Kind == kindSystem {
		return true, nil
	}
	if conversation.VisibilityPolicy == visibilityPublicRead && (conversation.Kind == kindWorld || conversation.Kind == kindCustom) {
		return true, nil
	}
	if !hasMember(conversation.MemberPlayerIDs, playerID) {
		return false, nil
	}

	switch conversation.Kind {
	case kindGuild:
		if s.guilds == nil {
			return true, nil
		}
		allowed, appErr := s.guilds.IsGuildMember(ctx, conversation.ResourceID, playerID)
		if appErr != nil {
			return false, appErr
		}
		return allowed, nil
	case kindParty:
		if s.parties == nil {
			return true, nil
		}
		allowed, appErr := s.parties.IsPartyMember(ctx, conversation.ResourceID, playerID)
		if appErr != nil {
			return false, appErr
		}
		return allowed, nil
	default:
		return true, nil
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

func isResourceBoundKind(kind string) bool {
	switch kind {
	case kindGuild, kindParty, kindWorld, kindSystem, kindCustom:
		return true
	default:
		return false
	}
}

func mergeMembers(existing []string, incoming []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(incoming))
	merged := make([]string, 0, len(existing)+len(incoming))
	for _, memberID := range existing {
		if _, ok := seen[memberID]; ok {
			continue
		}
		seen[memberID] = struct{}{}
		merged = append(merged, memberID)
	}
	for _, memberID := range incoming {
		if _, ok := seen[memberID]; ok {
			continue
		}
		seen[memberID] = struct{}{}
		merged = append(merged, memberID)
	}
	slices.Sort(merged)
	return merged
}

func buildChannelDescriptor(conversation domain.Conversation) domain.ChannelDescriptor {
	conversation = defaultConversationGovernance(conversation)
	descriptor := domain.ChannelDescriptor{
		ConversationID:   conversation.ID,
		Kind:             conversation.Kind,
		ResourceID:       conversation.ResourceID,
		Scope:            channelScopeDirect,
		MembershipMode:   membershipExplicit,
		SendPolicy:       conversation.SendPolicy,
		VisibilityPolicy: conversation.VisibilityPolicy,
		ModerationMode:   conversation.ModerationMode,
		ResourceRequired: false,
		MemberCount:      len(conversation.MemberPlayerIDs),
		ModeratorIDs:     append([]string(nil), conversation.ModeratorIDs...),
		MutedPlayerIDs:   append([]string(nil), conversation.MutedPlayerIDs...),
	}

	if isResourceBoundKind(conversation.Kind) {
		descriptor.Scope = channelScopeResource
		descriptor.MembershipMode = membershipBound
		descriptor.ResourceRequired = true
	}
	return descriptor
}

func (s *ChatService) buildConversationSummary(conversation domain.Conversation, playerID string) (domain.ConversationSummary, *apperrors.Error) {
	cursor, _, err := s.readCursors.GetCursor(conversation.ID, playerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.ConversationSummary{}, &internal
	}
	messages, err := s.messages.ListMessages(conversation.ID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.ConversationSummary{}, &internal
	}

	var lastMessage *domain.Message
	updatedAt := conversation.CreatedAt
	if len(messages) > 0 {
		msg := messages[len(messages)-1]
		lastMessage = &msg
		updatedAt = msg.CreatedAt
	}
	unread := conversation.LastSeq - cursor.AckSeq
	if unread < 0 {
		unread = 0
	}

	return domain.ConversationSummary{
		ConversationID: conversation.ID,
		Kind:           conversation.Kind,
		ResourceID:     conversation.ResourceID,
		PlayerID:       playerID,
		LastSeq:        conversation.LastSeq,
		AckSeq:         cursor.AckSeq,
		UnreadCount:    unread,
		LastMessage:    lastMessage,
		UpdatedAt:      updatedAt,
	}, nil
}

func (s *ChatService) String() string {
	conversations, err := s.conversations.ListConversations()
	if err != nil {
		return "chat-service(conversations=unknown)"
	}
	return fmt.Sprintf("chat-service(conversations=%d)", len(conversations))
}

func (s *ChatService) enqueueOfflineDeliveries(ctx context.Context, conversation domain.Conversation, message domain.Message, targets []DeliveryTarget) {
	for _, target := range targets {
		if target.DeliveryMode != deliveryModeReplay {
			continue
		}

		payload, err := json.Marshal(map[string]any{
			"conversation_id":   conversation.ID,
			"conversation_kind": conversation.Kind,
			"resource_id":       conversation.ResourceID,
			"message_id":        message.ID,
			"seq":               message.Seq,
			"sender_player_id":  message.SenderPlayerID,
			"recipient_player":  target.PlayerID,
			"delivery_mode":     target.DeliveryMode,
		})
		if err != nil {
			continue
		}

		_ = s.scheduler.EnqueueJob(ctx, offlineJobType, string(payload))
	}
}





