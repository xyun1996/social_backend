package service

import "testing"

func TestConversationSendAckAndReplay(t *testing.T) {
	t.Parallel()

	svc := NewChatService()
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

	svc := NewChatService()
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

	svc := NewChatService()
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
