package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

const (
	// JobsIndexKey is the Redis sorted-set index for worker jobs.
	JobsIndexKey = "worker:jobs"
	// JobKeyPrefix is the Redis prefix for per-job payloads.
	JobKeyPrefix = "worker:job:"
)

// Repository is the Redis repository for durable worker queue state.
type Repository struct {
	config db.RedisConfig
	client *redis.Client
}

// NewRepository constructs the worker Redis repository.
func NewRepository(config db.RedisConfig, client *redis.Client) *Repository {
	return &Repository{config: config, client: client}
}

// JobKey builds the canonical Redis key for a job id.
func (r *Repository) JobKey(jobID string) string {
	return JobKeyPrefix + jobID
}

// URL returns the shared Redis URL used by this repository.
func (r *Repository) URL() string {
	return r.config.URL()
}

func (r *Repository) String() string {
	return fmt.Sprintf("worker-redis-repository(%s)", r.config.URL())
}

// SaveJob persists a job snapshot and keeps it in the ordered index.
func (r *Repository) SaveJob(job domain.Job) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis repository is not configured")
	}

	raw, err := json.Marshal(job)
	if err != nil {
		return err
	}

	ctx := context.Background()
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, r.JobKey(job.ID), raw, 0)
	pipe.ZAdd(ctx, JobsIndexKey, redis.Z{
		Score:  float64(job.CreatedAt.UTC().UnixNano()),
		Member: job.ID,
	})
	_, err = pipe.Exec(ctx)
	return err
}

// GetJob loads a job snapshot by id.
func (r *Repository) GetJob(jobID string) (domain.Job, bool, error) {
	if r == nil || r.client == nil {
		return domain.Job{}, false, fmt.Errorf("redis repository is not configured")
	}

	raw, err := r.client.Get(context.Background(), r.JobKey(jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return domain.Job{}, false, nil
		}
		return domain.Job{}, false, err
	}

	var job domain.Job
	if err := json.Unmarshal(raw, &job); err != nil {
		return domain.Job{}, false, err
	}
	return job, true, nil
}

// ListJobs returns all persisted jobs ordered by created time then id.
func (r *Repository) ListJobs() ([]domain.Job, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("redis repository is not configured")
	}

	ctx := context.Background()
	ids, err := r.client.ZRange(ctx, JobsIndexKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	jobs := make([]domain.Job, 0, len(ids))
	for _, id := range ids {
		job, ok, err := r.GetJob(id)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		jobs = append(jobs, job)
	}

	sort.SliceStable(jobs, func(i, j int) bool {
		if !jobs[i].CreatedAt.Equal(jobs[j].CreatedAt) {
			return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
		}
		return jobs[i].ID < jobs[j].ID
	})
	return jobs, nil
}
