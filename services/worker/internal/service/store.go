package service

import (
	"slices"
	"sync"

	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

// JobStore persists worker queue state.
type JobStore interface {
	ListJobs() ([]domain.Job, error)
	SaveJob(job domain.Job) error
	GetJob(jobID string) (domain.Job, bool, error)
}

type memoryJobStore struct {
	mu   sync.RWMutex
	jobs map[string]domain.Job
}

func newMemoryJobStore() *memoryJobStore {
	return &memoryJobStore{
		jobs: make(map[string]domain.Job),
	}
}

func (s *memoryJobStore) ListJobs() ([]domain.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]domain.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	slices.SortFunc(jobs, func(a domain.Job, b domain.Job) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})
	return jobs, nil
}

func (s *memoryJobStore) SaveJob(job domain.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

func (s *memoryJobStore) GetJob(jobID string) (domain.Job, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[jobID]
	return job, ok, nil
}
