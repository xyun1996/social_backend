package service

import "testing"

func TestConnectHeartbeatDisconnectLifecycle(t *testing.T) {
	t.Parallel()

	svc := NewPresenceService()

	connected, err := svc.Connect("p1", "sess-1", "realm-1", "lobby")
	if err != nil {
		t.Fatalf("connect returned error: %+v", err)
	}

	if connected.Status != statusOnline {
		t.Fatalf("unexpected status after connect: %q", connected.Status)
	}

	heartbeat, err := svc.Heartbeat("p1", "sess-1", "realm-1", "queue")
	if err != nil {
		t.Fatalf("heartbeat returned error: %+v", err)
	}

	if heartbeat.Location != "queue" {
		t.Fatalf("unexpected location after heartbeat: %q", heartbeat.Location)
	}

	disconnected, err := svc.Disconnect("p1", "sess-1")
	if err != nil {
		t.Fatalf("disconnect returned error: %+v", err)
	}

	if disconnected.Status != statusOffline {
		t.Fatalf("unexpected status after disconnect: %q", disconnected.Status)
	}
}

func TestHeartbeatRejectsDifferentSession(t *testing.T) {
	t.Parallel()

	svc := NewPresenceService()
	if _, err := svc.Connect("p1", "sess-1", "", ""); err != nil {
		t.Fatalf("connect returned error: %+v", err)
	}

	if _, err := svc.Heartbeat("p1", "sess-2", "", ""); err == nil {
		t.Fatalf("expected heartbeat with different session to fail")
	}
}
