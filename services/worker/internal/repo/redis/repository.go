package redis

import (
	"fmt"

	"github.com/xyun1996/social_backend/pkg/db"
)

const (
	// QueueKeyPrefix is the Redis prefix for queued worker jobs.
	QueueKeyPrefix = "worker:queue:"
	// ClaimKeyPrefix is the Redis prefix for worker claim metadata.
	ClaimKeyPrefix = "worker:claim:"
)

// Repository is the Redis foundation for future worker queue state.
type Repository struct {
	config db.RedisConfig
}

// NewRepository constructs the worker Redis repository foundation.
func NewRepository(config db.RedisConfig) *Repository {
	return &Repository{config: config}
}

// QueueKey builds the canonical Redis queue key for a job type.
func (r *Repository) QueueKey(jobType string) string {
	return QueueKeyPrefix + jobType
}

// ClaimKey builds the canonical Redis claim key for a job id.
func (r *Repository) ClaimKey(jobID string) string {
	return ClaimKeyPrefix + jobID
}

// URL returns the shared Redis URL used by this repository.
func (r *Repository) URL() string {
	return r.config.URL()
}

func (r *Repository) String() string {
	return fmt.Sprintf("worker-redis-repository(%s)", r.config.URL())
}
