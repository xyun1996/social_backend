package service

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

type fakePresenceReader struct {
	snapshots map[string]PresenceSnapshot
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

	svc := NewChatService(nil)
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

	svc := NewChatService(nil)
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

	svc := NewChatService(nil)
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
	})

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
