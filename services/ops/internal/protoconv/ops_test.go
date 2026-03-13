package protoconv

import (
	"testing"

	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

func TestToProtoPlayerOverview(t *testing.T) {
	t.Parallel()

	got := ToProtoPlayerOverview(opsservice.PlayerOverview{
		PlayerID: "p1",
		Presence: opsservice.PresenceRecord{
			PlayerID:  "p1",
			Status:    "online",
			SessionID: "sess-1",
		},
		Friends:            []string{"p2"},
		Blocks:             []string{"p3"},
		PendingInbox:       []string{"p4"},
		PendingOutbox:      []string{"p5"},
		FriendCnt:          1,
		BlockCnt:           1,
		PendingInboxCount:  1,
		PendingOutboxCount: 1,
	})

	if got.GetPlayerId() != "p1" || got.GetPresence().GetSessionId() != "sess-1" {
		t.Fatalf("unexpected proto overview identity fields: %+v", got)
	}
	if got.GetFriendCount() != 1 || got.GetPendingOutboxCount() != 1 {
		t.Fatalf("unexpected proto overview counters: %+v", got)
	}
}

func TestToProtoDurableSummary(t *testing.T) {
	t.Parallel()

	got := ToProtoDurableSummary(opsservice.DurableSummary{
		MySQL: &opsservice.MySQLBootstrapSnapshot{
			Count: 1,
			Services: []opsservice.MySQLBootstrapService{
				{Service: "identity", Count: 1, MigrationIDs: []string{"001_identity_core"}},
			},
		},
		Redis: &opsservice.RedisRuntimeSnapshot{
			RedisURL:            "redis://localhost:6379/0",
			PresenceRecordCount: 1,
			WorkerStatusCounters: []opsservice.RedisWorkerStatusCount{
				{Status: "queued", Count: 2},
			},
		},
	})

	if got.GetMysql().GetCount() != 1 || got.GetMysql().GetServices()[0].GetService() != "identity" {
		t.Fatalf("unexpected mysql durable summary conversion: %+v", got)
	}
	if got.GetRedis().GetPresenceRecordCount() != 1 || got.GetRedis().GetWorkerStatusCounters()[0].GetCount() != 2 {
		t.Fatalf("unexpected redis durable summary conversion: %+v", got)
	}
}
