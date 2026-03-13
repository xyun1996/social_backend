package service

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

type fakePresenceReader struct {
	snapshots map[string]PresenceSnapshot
}

type fakeJobScheduler struct {
	jobTypes []string
	payloads []string
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
