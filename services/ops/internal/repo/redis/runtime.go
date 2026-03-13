package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

const (
	presenceKeyPattern = "presence:player:*"
	gatewaySessionsKey = "gateway:sessions"
	workerJobsIndexKey = "worker:jobs"
	workerJobKeyPrefix = "worker:job:"
)

type workerJobRecord struct {
	Status string `json:"status"`
}

// RuntimeReader reads operator-facing Redis runtime state.
type RuntimeReader struct {
	config db.RedisConfig
	client *redis.Client
}

// NewRuntimeReader constructs the ops Redis runtime reader.
func NewRuntimeReader(config db.RedisConfig, client *redis.Client) *RuntimeReader {
	return &RuntimeReader{config: config, client: client}
}

// GetRedisRuntimeSnapshot reads current Redis-backed runtime state.
func (r *RuntimeReader) GetRedisRuntimeSnapshot(ctx context.Context) (opsservice.RedisRuntimeSnapshot, *apperrors.Error) {
	if r == nil || r.client == nil {
		err := apperrors.New("dependency_missing", "redis runtime reader is not configured", 500)
		return opsservice.RedisRuntimeSnapshot{}, &err
	}

	presenceKeys, err := r.client.Keys(ctx, presenceKeyPattern).Result()
	if err != nil {
		appErr := apperrors.New("redis_query_failed", err.Error(), 500)
		return opsservice.RedisRuntimeSnapshot{}, &appErr
	}

	sessionCount, err := r.client.SCard(ctx, gatewaySessionsKey).Result()
	if err != nil {
		appErr := apperrors.New("redis_query_failed", err.Error(), 500)
		return opsservice.RedisRuntimeSnapshot{}, &appErr
	}

	jobIDs, err := r.client.ZRange(ctx, workerJobsIndexKey, 0, -1).Result()
	if err != nil {
		appErr := apperrors.New("redis_query_failed", err.Error(), 500)
		return opsservice.RedisRuntimeSnapshot{}, &appErr
	}

	statusCounts := map[string]int{}
	for _, jobID := range jobIDs {
		raw, err := r.client.Get(ctx, workerJobKeyPrefix+jobID).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			appErr := apperrors.New("redis_query_failed", err.Error(), 500)
			return opsservice.RedisRuntimeSnapshot{}, &appErr
		}

		var job workerJobRecord
		if err := json.Unmarshal(raw, &job); err != nil {
			appErr := apperrors.New("redis_query_failed", err.Error(), 500)
			return opsservice.RedisRuntimeSnapshot{}, &appErr
		}
		statusCounts[job.Status]++
	}

	statuses := make([]opsservice.RedisWorkerStatusCount, 0, len(statusCounts))
	for status, count := range statusCounts {
		statuses = append(statuses, opsservice.RedisWorkerStatusCount{Status: status, Count: count})
	}
	slices.SortFunc(statuses, func(a opsservice.RedisWorkerStatusCount, b opsservice.RedisWorkerStatusCount) int {
		switch {
		case a.Status < b.Status:
			return -1
		case a.Status > b.Status:
			return 1
		default:
			return 0
		}
	})

	return opsservice.RedisRuntimeSnapshot{
		RedisURL:             r.config.URL(),
		PresenceRecordCount:  len(presenceKeys),
		GatewaySessionCount:  int(sessionCount),
		WorkerJobCount:       len(jobIDs),
		WorkerStatusCounters: statuses,
	}, nil
}

func (r *RuntimeReader) String() string {
	return fmt.Sprintf("ops-redis-runtime-reader(%s)", r.config.URL())
}
