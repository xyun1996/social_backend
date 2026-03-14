package privatechat

import (
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
)

const defaultReplayLimit = 50

type Conversation struct {
	ID              string    `json:"id"`
	Kind            string    `json:"kind"`
	MemberPlayerIDs []string  `json:"member_player_ids"`
	LastSeq         int64     `json:"last_seq"`
	CreatedAt       time.Time `json:"created_at"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Seq            int64     `json:"seq"`
	SenderPlayerID string    `json:"sender_player_id"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

type Summary struct {
	ConversationID string    `json:"conversation_id"`
	PlayerID       string    `json:"player_id"`
	LastSeq        int64     `json:"last_seq"`
	AckSeq         int64     `json:"ack_seq"`
	UnreadCount    int64     `json:"unread_count"`
	LastMessage    *Message  `json:"last_message,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Service struct {
	mu               sync.RWMutex
	now              func() time.Time
	conversations    map[string]Conversation
	messages         map[string][]Message
	conversationByDM map[string]string
	readCursors      map[string]map[string]int64
}

func NewService() *Service {
	return &Service{
		now:              time.Now,
		conversations:    make(map[string]Conversation),
		messages:         make(map[string][]Message),
		conversationByDM: make(map[string]string),
		readCursors:      make(map[string]map[string]int64),
	}
}

func (s *Service) CreateConversation(memberPlayerIDs []string) (Conversation, *apperrors.Error) {
	members := normalizeMembers(memberPlayerIDs)
	if len(members) != 2 {
		err := apperrors.New("invalid_request", "private chat requires exactly 2 members", http.StatusBadRequest)
		return Conversation{}, &err
	}

	key := members[0] + ":" + members[1]

	s.mu.Lock()
	defer s.mu.Unlock()
	if conversationID, ok := s.conversationByDM[key]; ok {
		return s.conversations[conversationID], nil
	}

	conversationID, err := idgen.Token(8)
	if err != nil {
		internal := apperrors.Internal()
		return Conversation{}, &internal
	}

	record := Conversation{
		ID:              conversationID,
		Kind:            "private",
		MemberPlayerIDs: members,
		LastSeq:         0,
		CreatedAt:       s.now(),
	}
	s.conversations[conversationID] = record
	s.conversationByDM[key] = conversationID
	return record, nil
}

func (s *Service) ListConversations(playerID string) ([]Conversation, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]Conversation, 0)
	for _, conversation := range s.conversations {
		if slices.Contains(conversation.MemberPlayerIDs, playerID) {
			list = append(list, conversation)
		}
	}
	slices.SortFunc(list, func(a, b Conversation) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return list, nil
}

func (s *Service) SendMessage(conversationID, senderPlayerID, body string) (Message, *apperrors.Error) {
	if conversationID == "" || senderPlayerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and sender_player_id are required", http.StatusBadRequest)
		return Message{}, &err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		err := apperrors.New("invalid_request", "body is required", http.StatusBadRequest)
		return Message{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return Message{}, &err
	}
	if !slices.Contains(conversation.MemberPlayerIDs, senderPlayerID) {
		err := apperrors.New("forbidden", "sender is not in the conversation", http.StatusForbidden)
		return Message{}, &err
	}

	messageID, err := idgen.Token(10)
	if err != nil {
		internal := apperrors.Internal()
		return Message{}, &internal
	}
	conversation.LastSeq++
	s.conversations[conversationID] = conversation

	record := Message{
		ID:             messageID,
		ConversationID: conversationID,
		Seq:            conversation.LastSeq,
		SenderPlayerID: senderPlayerID,
		Body:           body,
		CreatedAt:      s.now(),
	}
	s.messages[conversationID] = append(s.messages[conversationID], record)
	return record, nil
}

func (s *Service) ReplayMessages(conversationID, playerID string, afterSeq int64, limit int) ([]Message, *apperrors.Error) {
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

	s.mu.RLock()
	defer s.mu.RUnlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return nil, &err
	}
	if !slices.Contains(conversation.MemberPlayerIDs, playerID) {
		err := apperrors.New("forbidden", "player is not in the conversation", http.StatusForbidden)
		return nil, &err
	}

	messages := s.messages[conversationID]
	result := make([]Message, 0, min(limit, len(messages)))
	for _, message := range messages {
		if message.Seq <= afterSeq {
			continue
		}
		result = append(result, message)
		if len(result) == limit {
			break
		}
	}
	return result, nil
}

func (s *Service) AckConversation(conversationID, playerID string, ackSeq int64) *apperrors.Error {
	if conversationID == "" || playerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and player_id are required", http.StatusBadRequest)
		return &err
	}
	if ackSeq < 0 {
		err := apperrors.New("invalid_request", "ack_seq must be >= 0", http.StatusBadRequest)
		return &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return &err
	}
	if !slices.Contains(conversation.MemberPlayerIDs, playerID) {
		err := apperrors.New("forbidden", "player is not in the conversation", http.StatusForbidden)
		return &err
	}
	if ackSeq > conversation.LastSeq {
		err := apperrors.New("invalid_request", "ack_seq cannot exceed last_seq", http.StatusBadRequest)
		return &err
	}
	if _, ok := s.readCursors[conversationID]; !ok {
		s.readCursors[conversationID] = make(map[string]int64)
	}
	if current := s.readCursors[conversationID][playerID]; ackSeq < current {
		ackSeq = current
	}
	s.readCursors[conversationID][playerID] = ackSeq
	return nil
}

func (s *Service) GetSummary(conversationID, playerID string) (Summary, *apperrors.Error) {
	if conversationID == "" || playerID == "" {
		err := apperrors.New("invalid_request", "conversation_id and player_id are required", http.StatusBadRequest)
		return Summary{}, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return Summary{}, &err
	}
	if !slices.Contains(conversation.MemberPlayerIDs, playerID) {
		err := apperrors.New("forbidden", "player is not in the conversation", http.StatusForbidden)
		return Summary{}, &err
	}

	ackSeq := s.readCursors[conversationID][playerID]
	unread := conversation.LastSeq - ackSeq
	if unread < 0 {
		unread = 0
	}
	var lastMessage *Message
	var updatedAt time.Time
	if items := s.messages[conversationID]; len(items) > 0 {
		msg := items[len(items)-1]
		lastMessage = &msg
		updatedAt = msg.CreatedAt
	} else {
		updatedAt = conversation.CreatedAt
	}
	return Summary{
		ConversationID: conversationID,
		PlayerID:       playerID,
		LastSeq:        conversation.LastSeq,
		AckSeq:         ackSeq,
		UnreadCount:    unread,
		LastMessage:    lastMessage,
		UpdatedAt:      updatedAt,
	}, nil
}

func (s *Service) ListSummaries(playerID string) ([]Summary, *apperrors.Error) {
	conversations, appErr := s.ListConversations(playerID)
	if appErr != nil {
		return nil, appErr
	}
	result := make([]Summary, 0, len(conversations))
	for _, conversation := range conversations {
		summary, appErr := s.GetSummary(conversation.ID, playerID)
		if appErr != nil {
			return nil, appErr
		}
		result = append(result, summary)
	}
	slices.SortFunc(result, func(a, b Summary) int {
		if !a.UpdatedAt.Equal(b.UpdatedAt) {
			if a.UpdatedAt.After(b.UpdatedAt) {
				return -1
			}
			return 1
		}
		if a.ConversationID < b.ConversationID {
			return -1
		}
		if a.ConversationID > b.ConversationID {
			return 1
		}
		return 0
	})
	return result, nil
}

func normalizeMembers(members []string) []string {
	seen := make(map[string]struct{}, len(members))
	result := make([]string, 0, len(members))
	for _, memberID := range members {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			continue
		}
		if _, ok := seen[memberID]; ok {
			continue
		}
		seen[memberID] = struct{}{}
		result = append(result, memberID)
	}
	slices.Sort(result)
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
