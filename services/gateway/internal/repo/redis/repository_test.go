package redis

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

func TestRepositoryRoundTripSessionAndEvents(t *testing.T) {
	t.Parallel()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	defer server.Close()

	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer client.Close()

	repo := NewRepository(db.RedisConfig{Addr: server.Addr()}, client)
	session := gatewayservice.RealtimeSession{
		SessionID: "sess-1",
		AccountID: "a1",
		PlayerID:  "p1",
		State:     "active",
	}

	if err := repo.SaveSession(session); err != nil {
		t.Fatalf("SaveSession returned error: %v", err)
	}

	loaded, ok, err := repo.GetSession(session.SessionID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if !ok || loaded.SessionID != session.SessionID {
		t.Fatalf("unexpected loaded session: %+v", loaded)
	}

	sessions, err := repo.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].SessionID != session.SessionID {
		t.Fatalf("unexpected sessions: %+v", sessions)
	}

	events := []gatewayservice.ChatMessageEnvelope{{EventID: "evt-1"}, {EventID: "evt-2"}}
	if err := repo.SaveEvents(session.SessionID, events); err != nil {
		t.Fatalf("SaveEvents returned error: %v", err)
	}

	loadedEvents, err := repo.GetEvents(session.SessionID)
	if err != nil {
		t.Fatalf("GetEvents returned error: %v", err)
	}
	if len(loadedEvents) != 2 || loadedEvents[0].EventID != "evt-1" {
		t.Fatalf("unexpected loaded events: %+v", loadedEvents)
	}
}
