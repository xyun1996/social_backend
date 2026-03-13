package redis

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
)

func TestGetRedisRuntimeSnapshot(t *testing.T) {
	t.Parallel()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	defer server.Close()

	config := db.RedisConfig{Addr: server.Addr(), DB: 0}
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer client.Close()

	if err := client.Set(context.Background(), "presence:player:p1", `{"status":"online"}`, 0).Err(); err != nil {
		t.Fatalf("seed presence: %v", err)
	}
	if err := client.SAdd(context.Background(), gatewaySessionsKey, "sess-1").Err(); err != nil {
		t.Fatalf("seed gateway session: %v", err)
	}
	if err := client.ZAdd(context.Background(), workerJobsIndexKey,
		redis.Z{Score: 1, Member: "job-1"},
		redis.Z{Score: 2, Member: "job-2"},
	).Err(); err != nil {
		t.Fatalf("seed worker index: %v", err)
	}
	if err := client.Set(context.Background(), workerJobKeyPrefix+"job-1", `{"status":"queued"}`, 0).Err(); err != nil {
		t.Fatalf("seed worker job 1: %v", err)
	}
	if err := client.Set(context.Background(), workerJobKeyPrefix+"job-2", `{"status":"completed"}`, 0).Err(); err != nil {
		t.Fatalf("seed worker job 2: %v", err)
	}

	reader := NewRuntimeReader(config, client)
	snapshot, appErr := reader.GetRedisRuntimeSnapshot(context.Background())
	if appErr != nil {
		t.Fatalf("GetRedisRuntimeSnapshot returned error: %+v", appErr)
	}
	if snapshot.PresenceRecordCount != 1 || snapshot.GatewaySessionCount != 1 || snapshot.WorkerJobCount != 2 {
		t.Fatalf("unexpected redis runtime snapshot: %+v", snapshot)
	}
	if len(snapshot.WorkerStatusCounters) != 2 {
		t.Fatalf("unexpected worker status counters: %+v", snapshot.WorkerStatusCounters)
	}
}
