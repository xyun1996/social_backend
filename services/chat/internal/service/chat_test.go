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

type fakeGuildMembershipReader struct {
	members map[string]map[string]bool
}

type fakePartyMembershipReader struct {
	members map[string]map[string]bool
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

func (f *fakeGuildMembershipReader) IsGuildMember(_ context.Context, guildID string, playerID string) (bool, *apperrors.Error) {
	return f.members[guildID][playerID], nil
}

func (f *fakePartyMembershipReader) IsPartyMember(_ context.Context, partyID string, playerID string) (bool, *apperrors.Error) {
	return f.members[partyID][playerID], nil
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

func TestResourceBackedConversationRequiresResourceAndReusesChannel(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)

	if _, err := svc.CreateConversation(kindGuild, "", []string{"p1"}); err == nil {
		t.Fatalf("expected guild conversation without resource_id to fail")
	}

	first, err := svc.CreateConversation(kindGuild, "guild-1", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create guild conversation returned error: %+v", err)
	}
	second, err := svc.CreateConversation(kindGuild, "guild-1", []string{"p2", "p3"})
	if err != nil {
		t.Fatalf("recreate guild conversation returned error: %+v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected resource-backed channel reuse, got first=%s second=%s", first.ID, second.ID)
	}
	if len(second.MemberPlayerIDs) != 3 || second.MemberPlayerIDs[2] != "p3" {
		t.Fatalf("expected resource-backed channel members to reconcile, got %+v", second.MemberPlayerIDs)
	}
}

func TestGuildConversationSendRequiresCurrentGuildMembership(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	svc.SetMembershipReaders(&fakeGuildMembershipReader{
		members: map[string]map[string]bool{
			"guild-1": {
				"p1": true,
				"p2": false,
			},
		},
	}, nil)

	conversation, err := svc.CreateConversation(kindGuild, "guild-1", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create guild conversation returned error: %+v", err)
	}

	if _, sendErr := svc.SendMessage(conversation.ID, "p2", "hello"); sendErr == nil {
		t.Fatalf("expected stale guild member send to be rejected")
	}
	if _, sendErr := svc.SendMessage(conversation.ID, "p1", "hello"); sendErr != nil {
		t.Fatalf("expected current guild member send to succeed: %+v", sendErr)
	}
}

func TestPartyConversationVisibilityRequiresCurrentPartyMembership(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	svc.SetMembershipReaders(nil, &fakePartyMembershipReader{
		members: map[string]map[string]bool{
			"party-1": {
				"p1": true,
				"p2": false,
			},
		},
	})

	conversation, err := svc.CreateConversation(kindParty, "party-1", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create party conversation returned error: %+v", err)
	}
	if _, err := svc.SendMessage(conversation.ID, "p1", "hello"); err != nil {
		t.Fatalf("send message returned error: %+v", err)
	}

	conversations, listErr := svc.ListConversations("p2")
	if listErr != nil {
		t.Fatalf("list conversations returned error: %+v", listErr)
	}
	if len(conversations) != 0 {
		t.Fatalf("expected stale party member to lose visibility: %+v", conversations)
	}

	if _, replayErr := svc.ReplayMessages(conversation.ID, "p2", 0, 10); replayErr == nil {
		t.Fatalf("expected replay to reject stale party member")
	}
}

func TestGetChannelDescriptorExplainsChannelPolicy(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	conversation, err := svc.CreateConversation(kindSystem, "system-global", []string{"p1"})
	if err != nil {
		t.Fatalf("create system conversation returned error: %+v", err)
	}

	descriptor, descErr := svc.GetChannelDescriptor(conversation.ID)
	if descErr != nil {
		t.Fatalf("get channel descriptor returned error: %+v", descErr)
	}
	if descriptor.Scope != channelScopeResource || descriptor.MembershipMode != membershipBound {
		t.Fatalf("unexpected channel descriptor scope: %+v", descriptor)
	}
	if descriptor.SendPolicy != sendPolicySystemOnly || !descriptor.ResourceRequired {
		t.Fatalf("unexpected channel descriptor policy: %+v", descriptor)
	}
}

func TestConversationSummaryTracksUnreadAndLastMessage(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	conversation, err := svc.CreateConversation(kindPrivate, "", []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("create conversation returned error: %+v", err)
	}
	if _, err := svc.SendMessage(conversation.ID, "p1", "hello"); err != nil {
		t.Fatalf("send first message returned error: %+v", err)
	}
	if _, err := svc.SendMessage(conversation.ID, "p2", "hi"); err != nil {
		t.Fatalf("send second message returned error: %+v", err)
	}
	if _, err := svc.AckConversation(conversation.ID, "p2", 1); err != nil {
		t.Fatalf("ack returned error: %+v", err)
	}

	summary, summaryErr := svc.GetConversationSummary(conversation.ID, "p2")
	if summaryErr != nil {
		t.Fatalf("get conversation summary returned error: %+v", summaryErr)
	}
	if summary.UnreadCount != 1 || summary.LastMessage == nil || summary.LastMessage.Body != "hi" {
		t.Fatalf("unexpected conversation summary: %+v", summary)
	}

	summaries, listErr := svc.ListConversationSummaries("p2")
	if listErr != nil {
		t.Fatalf("list conversation summaries returned error: %+v", listErr)
	}
	if len(summaries) != 1 || summaries[0].ConversationID != conversation.ID {
		t.Fatalf("unexpected conversation summary list: %+v", summaries)
	}
}

func TestGroupConversationRequiresTwoMembers(t *testing.T) {
	t.Parallel()

	svc := NewChatService(nil, nil)
	if _, err := svc.CreateConversation(kindGroup, "", []string{"p1"}); err == nil {
		t.Fatalf("expected group conversation with one member to fail")
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
