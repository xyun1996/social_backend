package redis

import (
	"testing"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/presence/internal/domain"
)

func TestKeyUsesCanonicalPrefix(t *testing.T) {
	t.Parallel()

	store := NewStore(db.RedisConfig{Addr: "localhost:6379"}, nil)
	if key := store.Key("p1"); key != "presence:player:p1" {
		t.Fatalf("unexpected redis key: %s", key)
	}
}

func TestMarshalRoundTripPresenceSnapshot(t *testing.T) {
	t.Parallel()

	now := time.Unix(1700000000, 0).UTC()
	disconnected := now.Add(time.Minute)
	store := NewStore(db.RedisConfig{Addr: "localhost:6379"}, nil)

	raw, err := store.Marshal(domain.Presence{
		PlayerID:        "p1",
		Status:          "offline",
		SessionID:       "sess-1",
		RealmID:         "realm-1",
		Location:        "lobby",
		LastHeartbeatAt: now,
		LastSeenAt:      now,
		ConnectedAt:     &now,
		DisconnectedAt:  &disconnected,
	})
	if err != nil {
		t.Fatalf("marshal returned error: %v", err)
	}

	presence, err := store.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal returned error: %v", err)
	}
	if presence.PlayerID != "p1" || presence.Status != "offline" || presence.SessionID != "sess-1" {
		t.Fatalf("unexpected presence roundtrip: %+v", presence)
	}
	if presence.ConnectedAt == nil || presence.DisconnectedAt == nil {
		t.Fatalf("expected timestamp pointers to roundtrip: %+v", presence)
	}
}
