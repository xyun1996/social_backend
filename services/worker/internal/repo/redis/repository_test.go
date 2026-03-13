package redis

import (
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

func TestRepositorySaveAndListJobs(t *testing.T) {
	t.Parallel()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	defer server.Close()

	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer client.Close()

	repo := NewRepository(db.RedisConfig{Addr: server.Addr()}, client)
	first := domain.Job{
		ID:        "job-1",
		Type:      "invite.expire",
		Payload:   `{}`,
		Status:    "queued",
		CreatedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	}
	second := domain.Job{
		ID:        "job-2",
		Type:      "chat.offline_delivery",
		Payload:   `{}`,
		Status:    "queued",
		CreatedAt: first.CreatedAt.Add(time.Minute),
	}

	if err := repo.SaveJob(second); err != nil {
		t.Fatalf("SaveJob(second) returned error: %v", err)
	}
	if err := repo.SaveJob(first); err != nil {
		t.Fatalf("SaveJob(first) returned error: %v", err)
	}

	loaded, ok, err := repo.GetJob(first.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}
	if !ok || loaded.ID != first.ID {
		t.Fatalf("unexpected loaded job: %+v (ok=%v)", loaded, ok)
	}

	jobs, err := repo.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs returned error: %v", err)
	}
	if len(jobs) != 2 || jobs[0].ID != "job-1" || jobs[1].ID != "job-2" {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
}
