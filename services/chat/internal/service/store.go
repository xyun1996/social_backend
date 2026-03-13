package service

import (
	"sync"

	"github.com/xyun1996/social_backend/services/chat/internal/domain"
)

// ConversationStore persists conversation metadata and membership scope.
type ConversationStore interface {
	ListConversations() ([]domain.Conversation, error)
	SaveConversation(conversation domain.Conversation) error
	GetConversation(conversationID string) (domain.Conversation, bool, error)
}

// MessageStore persists ordered message history.
type MessageStore interface {
	ListMessages(conversationID string) ([]domain.Message, error)
	AppendMessage(message domain.Message) error
}

// ReadCursorStore persists per-player ack cursors.
type ReadCursorStore interface {
	GetCursor(conversationID string, playerID string) (domain.ReadCursor, bool, error)
	SaveCursor(cursor domain.ReadCursor) error
}

type memoryConversationStore struct {
	mu            sync.RWMutex
	conversations map[string]domain.Conversation
}

func newMemoryConversationStore() *memoryConversationStore {
	return &memoryConversationStore{
		conversations: make(map[string]domain.Conversation),
	}
}

func (s *memoryConversationStore) ListConversations() ([]domain.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conversations := make([]domain.Conversation, 0, len(s.conversations))
	for _, conversation := range s.conversations {
		conversations = append(conversations, conversation)
	}
	return conversations, nil
}

func (s *memoryConversationStore) SaveConversation(conversation domain.Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conversations[conversation.ID] = conversation
	return nil
}

func (s *memoryConversationStore) GetConversation(conversationID string) (domain.Conversation, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conversation, ok := s.conversations[conversationID]
	return conversation, ok, nil
}

type memoryMessageStore struct {
	mu       sync.RWMutex
	messages map[string][]domain.Message
}

func newMemoryMessageStore() *memoryMessageStore {
	return &memoryMessageStore{
		messages: make(map[string][]domain.Message),
	}
}

func (s *memoryMessageStore) ListMessages(conversationID string) ([]domain.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Message(nil), s.messages[conversationID]...), nil
}

func (s *memoryMessageStore) AppendMessage(message domain.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[message.ConversationID] = append(s.messages[message.ConversationID], message)
	return nil
}

type memoryReadCursorStore struct {
	mu      sync.RWMutex
	cursors map[string]map[string]domain.ReadCursor
}

func newMemoryReadCursorStore() *memoryReadCursorStore {
	return &memoryReadCursorStore{
		cursors: make(map[string]map[string]domain.ReadCursor),
	}
}

func (s *memoryReadCursorStore) GetCursor(conversationID string, playerID string) (domain.ReadCursor, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	playerCursors := s.cursors[conversationID]
	if playerCursors == nil {
		return domain.ReadCursor{}, false, nil
	}
	cursor, ok := playerCursors[playerID]
	return cursor, ok, nil
}

func (s *memoryReadCursorStore) SaveCursor(cursor domain.ReadCursor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cursors[cursor.ConversationID] == nil {
		s.cursors[cursor.ConversationID] = make(map[string]domain.ReadCursor)
	}
	s.cursors[cursor.ConversationID][cursor.PlayerID] = cursor
	return nil
}
