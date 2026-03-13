package service

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/chat/internal/domain"
)

type fakePresenceReader struct {
	snapshots map[string]PresenceSnapshot
}

type fakeJobScheduler struct {
	jobTypes []string
	payloads []string
}

type recordingConversationStore struct {
	conversations map[string]domain.Conversation
}

func newRecordingConversationStore() *recordingConversationStore {
	return &recordingConversationStore{conversations: make(map[string]domain.Conversation)}
}

func (s *recordingConversationStore) ListConversations() ([]domain.Conversation, error) {
	conversations := make([]domain.Conversation, 0, len(s.conversations))
	for _, conversation := range s.conversations {
		conversations = append(conversations, conversation)
	}
	return conversations, nil
}

func (s *recordingConversationStore) SaveConversation(conversation domain.Conversation) error {
	s.conversations[conversation.ID] = conversation
	return nil
}

func (s *recordingConversationStore) GetConversation(conversationID string) (domain.Conversation, bool, error) {
	conversation, ok := s.conversations[conversationID]
	return conversation, ok, nil
}

type recordingMessageStore struct {
	messages map[string][]domain.Message
}

func newRecordingMessageStore() *recordingMessageStore {
	return &recordingMessageStore{messages: make(map[string][]domain.Message)}
}

func (s *recordingMessageStore) ListMessages(conversationID string) ([]domain.Message, error) {
	return append([]domain.Message(nil), s.messages[conversationID]...), nil
}

func (s *recordingMessageStore) AppendMessage(message domain.Message) error {
	s.messages[message.ConversationID] = append(s.messages[message.ConversationID], message)
	return nil
}

type recordingReadCursorStore struct {
	cursors map[string]map[string]domain.ReadCursor
}

func newRecordingReadCursorStore() *recordingReadCursorStore {
	return &recordingReadCursorStore{cursors: make(map[string]map[string]domain.ReadCursor)}
}

func (s *recordingReadCursorStore) GetCursor(conversationID string, playerID string) (domain.ReadCursor, bool, error) {
	if s.cursors[conversationID] == nil {
		return domain.ReadCursor{}, false, nil
	}
	cursor, ok := s.cursors[conversationID][playerID]
	return cursor, ok, nil
}

func (s *recordingReadCursorStore) SaveCursor(cursor domain.ReadCursor) error {
	if s.cursors[cursor.ConversationID] == nil {
		s.cursors[cursor.ConversationID] = make(map[string]domain.ReadCursor)
	}
	s.cursors[cursor.ConversationID][cursor.PlayerID] = cursor
	return nil
}

func (f *fakeJobScheduler) EnqueueJob(_ context.Context, jobType string, payload string) *apperrors.Error {
	f.jobTypes = append(f.jobTypes, jobType)
	f.payloads = append(f.payloads, payload)
	return nil
}

func (f *fakePresenceReader) GetPresence(_ context.Context, playerID string) (PresenceSnapshot, *apperrors.Error) {
	snapshot, ok := f.snapshots[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", 404)
		return PresenceSnapshot{}, &err
	}

	return snapshot, nil
}

func TestConversationSendAckAndReplay(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}

	first, sendErr := svc.SendMessage(conversation.ID, "p1", "hello")
	if sendErr != nil {
		t.Fatalf("send first message returned error: %+v", sendErr)
	}

	second, sendErr := svc.SendMessage(conversation.ID, "p2", "hi")
	if sendErr != nil {
		t.Fatalf("send second message returned error: %+v", sendErr)
	}

	if first.Seq != 1 || second.Seq != 2 {
		t.Fatalf("unexpected seq values: first=%d second=%d", first.Seq, second.Seq)
	}

	cursor, ackErr := svc.AckConversation(conversation.ID, "p2", 2)
	if ackErr != nil {
		t.Fatalf("ack returned error: %+v", ackErr)
	}

	if cursor.AckSeq != 2 {
		t.Fatalf("unexpected ack seq: %d", cursor.AckSeq)
	}

	replay, replayErr := svc.ReplayMessages(conversation.ID, "p2", 1, 50)
	if replayErr != nil {
		t.Fatalf("replay returned error: %+v", replayErr)
	}

	if len(replay) != 1 || replay[0].Seq != 2 {
		t.Fatalf("unexpected replay payload: %+v", replay)
	}
}

func TestPrivateConversationRequiresMembershipToSend(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}

	if _, sendErr := svc.SendMessage(conversation.ID, "p3", "oops"); sendErr == nil {
		t.Fatalf("expected non-member send to be rejected")
	}
}

func TestSystemConversationOnlyAllowsSystemSender(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	conversation, err := svc.CreateConversation(kindSystem, "system-global", []string{"p1"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}

	if _, sendErr := svc.SendMessage(conversation.ID, "p1", "oops"); sendErr == nil {
		t.Fatalf("expected player send to system conversation to be rejected")
	}

	if _, sendErr := svc.SendMessage(conversation.ID, "system", "maintenance"); sendErr != nil {
		t.Fatalf("expected system sender to be allowed: %+v", sendErr)
	}
}

func TestPlanDeliveryUsesPresenceForRouting(t *testing.T) {
	t.Parallel()

	svc := NewChatService(&fakePresenceReader{
		snapshots: map[string]PresenceSnapshot{
			"p2": {
				PlayerID:  "p2",
				Status:    presenceOnline,
				SessionID: "sess-2",
				RealmID:   "realm-1",
				Location:  "lobby",
			},
		},
	}, nil)

	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}

	targets, planErr := svc.PlanDelivery(context.Background(), conversation.ID, "p1")
	if planErr != nil {
		t.Fatalf("plan delivery returned error: %+v", planErr)
	}

	if len(targets) != 1 {
		t.Fatalf("unexpected delivery target count: %d", len(targets))
	}

	if targets[0].DeliveryMode != deliveryModePush {
		t.Fatalf("unexpected delivery mode: %+v", targets[0])
	}
}

func TestSendMessageEnqueuesOfflineDeliveryJobs(t *testing.T) {
	t.Parallel()

	scheduler := &fakeJobScheduler{}
	svc := NewChatService(&fakePresenceReader{
		snapshots: map[string]PresenceSnapshot{
			"p2": {
				PlayerID: "p2",
				Status:   presenceOffline,
			},
		},
	}, scheduler)

	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}

	message, sendErr := svc.SendMessage(conversation.ID, "p1", "hello offline")
	if sendErr != nil {
		t.Fatalf("send message returned error: %+v", sendErr)
	}

	if message.Seq != 1 {
		t.Fatalf("unexpected message seq: %d", message.Seq)
	}

	if len(scheduler.jobTypes) != 1 {
		t.Fatalf("expected one offline delivery job, got %d", len(scheduler.jobTypes))
	}

	if scheduler.jobTypes[0] != offlineJobType {
		t.Fatalf("unexpected job type: %q", scheduler.jobTypes[0])
	}
}

func TestSendMessageDoesNotEnqueueOfflineJobsForOnlineRecipients(t *testing.T) {
	t.Parallel()

	scheduler := &fakeJobScheduler{}
	svc := NewChatService(&fakePresenceReader{
		snapshots: map[string]PresenceSnapshot{
			"p2": {
				PlayerID:  "p2",
				Status:    presenceOnline,
				SessionID: "sess-2",
			},
		},
	}, scheduler)

	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}

	if _, sendErr := svc.SendMessage(conversation.ID, "p1", "hello online"); sendErr != nil {
		t.Fatalf("send message returned error: %+v", sendErr)
	}

	if len(scheduler.jobTypes) != 0 {
		t.Fatalf("expected no offline delivery jobs, got %d", len(scheduler.jobTypes))
	}
}

func TestRecordOfflineDelivery(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}

	message, sendErr := svc.SendMessage(conversation.ID, "p1", "hello")
	if sendErr != nil {
		t.Fatalf("send message returned error: %+v", sendErr)
	}

	receipt, recordErr := svc.RecordOfflineDelivery(map[string]any{
		"conversation_id":  conversation.ID,
		"message_id":       message.ID,
		"recipient_player": "p2",
		"delivery_mode":    deliveryModeReplay,
	})
	if recordErr != nil {
		t.Fatalf("record offline delivery returned error: %+v", recordErr)
	}
	if receipt.MessageID != message.ID {
		t.Fatalf("unexpected receipt: %+v", receipt)
	}
}

func TestChatServiceUsesInjectedStores(t *testing.T) {
	t.Parallel()

	conversations := newRecordingConversationStore()
	messages := newRecordingMessageStore()
	cursors := newRecordingReadCursorStore()
	svc := NewChatServiceWithStores(conversations, messages, cursors, nil, nil)

	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}
	message, err := svc.SendMessage(conversation.ID, "p1", "hello")
	if err != nil {
		t.Fatalf("send message returned error: %+v", err)
	}
	cursor, err := svc.AckConversation(conversation.ID, "p2", 1)
	if err != nil {
		t.Fatalf("ack returned error: %+v", err)
	}

	if _, ok := conversations.conversations[conversation.ID]; !ok {
		t.Fatalf("expected conversation to be saved in injected store")
	}
	if len(messages.messages[conversation.ID]) != 1 || messages.messages[conversation.ID][0].ID != message.ID {
		t.Fatalf("expected message to be saved in injected store: %+v", messages.messages)
	}
	if cursors.cursors[conversation.ID]["p2"].AckSeq != cursor.AckSeq {
		t.Fatalf("expected cursor to be saved in injected store: %+v", cursors.cursors)
	}
}
