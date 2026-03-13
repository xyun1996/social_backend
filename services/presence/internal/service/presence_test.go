package service

import (
	"testing"

	"github.com/xyun1996/social_backend/services/presence/internal/domain"
)

type recordingPresenceStore struct {
	presences map[string]storedPresence
}

type storedPresence struct {
	playerID string
	status   string
}

func newRecordingPresenceStore() *recordingPresenceStore {
	return &recordingPresenceStore{
		presences: make(map[string]storedPresence),
	}
}

func (s *recordingPresenceStore) SavePresence(presence domain.Presence) error {
	s.presences[presence.PlayerID] = storedPresence{
		playerID: presence.PlayerID,
		status:   presence.Status,
	}
	return nil
}

func (s *recordingPresenceStore) GetPresence(playerID string) (domain.Presence, bool, error) {
	record, ok := s.presences[playerID]
	if !ok {
		return domain.Presence{}, false, nil
	}
	return domain.Presence{
		PlayerID: record.playerID,
		Status:   record.status,
	}, true, nil
}

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

func TestConnectUsesInjectedStore(t *testing.T) {
	t.Parallel()

	store := newRecordingPresenceStore()
	svc := NewPresenceServiceWithStore(store)

	presence, err := svc.Connect("p9", "sess-9", "", "")
	if err != nil {
		t.Fatalf("connect returned error: %+v", err)
	}

	record, ok := store.presences["p9"]
	if !ok {
		t.Fatalf("expected presence to be stored")
	}
	if record.playerID != "p9" || record.status != statusOnline || presence.PlayerID != "p9" {
		t.Fatalf("unexpected stored presence: %+v", record)
	}
}
