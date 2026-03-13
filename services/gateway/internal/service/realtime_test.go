package service

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

type fakeIntrospector struct {
	subject Subject
	err     *apperrors.Error
}

func (f fakeIntrospector) Introspect(context.Context, string) (Subject, *apperrors.Error) {
	return f.subject, f.err
}

type fakePresence struct {
	snapshot PresenceSnapshot
	update   PresenceUpdate
}

func (f *fakePresence) Connect(_ context.Context, update PresenceUpdate) (PresenceSnapshot, *apperrors.Error) {
	f.update = update
	return f.snapshot, nil
}

func (f *fakePresence) Heartbeat(_ context.Context, update PresenceUpdate) (PresenceSnapshot, *apperrors.Error) {
	f.update = update
	return f.snapshot, nil
}

func (f *fakePresence) Disconnect(_ context.Context, update PresenceUpdate) (PresenceSnapshot, *apperrors.Error) {
	f.update = update
	return f.snapshot, nil
}

func TestRealtimeHandshakeResumeAndClose(t *testing.T) {
	t.Parallel()

	presence := &fakePresence{snapshot: PresenceSnapshot{Status: "online", LastHeartbeatAt: "2026-03-13T10:00:00Z"}}
	svc := NewRealtimeService(fakeIntrospector{subject: Subject{AccountID: "a1", PlayerID: "p1"}}, presence)

	session, err := svc.Handshake(context.Background(), HandshakeRequest{
		AccessToken:   "token-1",
		SessionID:     "sess-1",
		RealmID:       "realm-1",
		Location:      "lobby",
		ClientVersion: "dev",
	})
	if err != nil {
		t.Fatalf("handshake returned error: %+v", err)
	}
	if session.State != sessionStateActive || session.PlayerID != "p1" {
		t.Fatalf("unexpected handshake session: %+v", session)
	}

	heartbeated, err := svc.Heartbeat(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("heartbeat returned error: %+v", err)
	}
	if heartbeated.State != sessionStateActive {
		t.Fatalf("unexpected heartbeat session: %+v", heartbeated)
	}

	resumed, err := svc.Resume(context.Background(), ResumeRequest{
		AccessToken:       "token-1",
		SessionID:         "sess-1",
		LastServerEventID: "evt-42",
	})
	if err != nil {
		t.Fatalf("resume returned error: %+v", err)
	}
	if resumed.LastServerEventID != "evt-42" {
		t.Fatalf("unexpected resumed session: %+v", resumed)
	}

	closed, err := svc.Close(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("close returned error: %+v", err)
	}
	if closed.State != sessionStateClosed {
		t.Fatalf("unexpected closed session: %+v", closed)
	}
}

func TestEnqueueChatEvent(t *testing.T) {
	t.Parallel()

	presence := &fakePresence{snapshot: PresenceSnapshot{Status: "online", LastHeartbeatAt: "2026-03-13T10:00:00Z"}}
	svc := NewRealtimeService(fakeIntrospector{subject: Subject{AccountID: "a1", PlayerID: "p1"}}, presence)
	if _, err := svc.Handshake(context.Background(), HandshakeRequest{
		AccessToken: "token-1",
		SessionID:   "sess-1",
	}); err != nil {
		t.Fatalf("handshake returned error: %+v", err)
	}

	if appErr := svc.EnqueueChatEvent("sess-1", ChatMessageEnvelope{EventID: "evt-1", MessageID: "msg-1"}); appErr != nil {
		t.Fatalf("enqueue event returned error: %+v", appErr)
	}

	inbox, err := svc.GetSessionEvents("sess-1")
	if err != nil {
		t.Fatalf("get session events returned error: %+v", err)
	}
	if inbox.Count != 1 || inbox.Events[0].EventID != "evt-1" {
		t.Fatalf("unexpected inbox: %+v", inbox)
	}
}

func TestAcknowledgeConversationPrunesInbox(t *testing.T) {
	t.Parallel()

	presence := &fakePresence{snapshot: PresenceSnapshot{Status: "online", LastHeartbeatAt: "2026-03-13T10:00:00Z"}}
	svc := NewRealtimeService(fakeIntrospector{subject: Subject{AccountID: "a1", PlayerID: "p1"}}, presence)
	if _, err := svc.Handshake(context.Background(), HandshakeRequest{
		AccessToken: "token-1",
		SessionID:   "sess-1",
	}); err != nil {
		t.Fatalf("handshake returned error: %+v", err)
	}

	_ = svc.EnqueueChatEvent("sess-1", ChatMessageEnvelope{EventID: "msg-1:1", Stream: "chat", Kind: "chat.message", ConversationID: "conv-1", Seq: 1})
	_ = svc.EnqueueChatEvent("sess-1", ChatMessageEnvelope{EventID: "msg-2:2", Stream: "chat", Kind: "chat.message", ConversationID: "conv-1", Seq: 2})
	_ = svc.EnqueueChatEvent("sess-1", ChatMessageEnvelope{EventID: "msg-3:1", Stream: "chat", Kind: "chat.message", ConversationID: "conv-2", Seq: 1})

	result, err := svc.AcknowledgeConversation("sess-1", "conv-1", 1)
	if err != nil {
		t.Fatalf("acknowledge conversation returned error: %+v", err)
	}
	if result.PrunedCount != 1 || result.LastAckedEventID != "msg-1:1" {
		t.Fatalf("unexpected compaction result: %+v", result)
	}

	inbox, err := svc.GetSessionEvents("sess-1")
	if err != nil {
		t.Fatalf("get session events returned error: %+v", err)
	}
	if inbox.Count != 2 || inbox.Events[0].EventID != "msg-2:2" {
		t.Fatalf("unexpected inbox after ack: %+v", inbox)
	}
}

func TestResumeTrimsEventsThroughLastServerEventID(t *testing.T) {
	t.Parallel()

	presence := &fakePresence{snapshot: PresenceSnapshot{Status: "online", LastHeartbeatAt: "2026-03-13T10:00:00Z"}}
	svc := NewRealtimeService(fakeIntrospector{subject: Subject{AccountID: "a1", PlayerID: "p1"}}, presence)
	if _, err := svc.Handshake(context.Background(), HandshakeRequest{
		AccessToken: "token-1",
		SessionID:   "sess-1",
	}); err != nil {
		t.Fatalf("handshake returned error: %+v", err)
	}

	_ = svc.EnqueueChatEvent("sess-1", ChatMessageEnvelope{EventID: "evt-1", Stream: "chat", Kind: "chat.message", ConversationID: "conv-1", Seq: 1})
	_ = svc.EnqueueChatEvent("sess-1", ChatMessageEnvelope{EventID: "evt-2", Stream: "chat", Kind: "chat.message", ConversationID: "conv-1", Seq: 2})
	_ = svc.EnqueueChatEvent("sess-1", ChatMessageEnvelope{EventID: "evt-3", Stream: "chat", Kind: "chat.message", ConversationID: "conv-1", Seq: 3})

	_, err := svc.Resume(context.Background(), ResumeRequest{
		AccessToken:       "token-1",
		SessionID:         "sess-1",
		LastServerEventID: "evt-2",
	})
	if err != nil {
		t.Fatalf("resume returned error: %+v", err)
	}

	inbox, err := svc.GetSessionEvents("sess-1")
	if err != nil {
		t.Fatalf("get session events returned error: %+v", err)
	}
	if inbox.Count != 1 || inbox.Events[0].EventID != "evt-3" {
		t.Fatalf("unexpected inbox after resume trim: %+v", inbox)
	}
}
