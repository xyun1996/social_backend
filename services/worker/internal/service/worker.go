package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

// JobHandler executes a claimed worker job.
type JobHandler func(ctx context.Context, job domain.Job) *apperrors.Error

// ExecutionResult summarizes worker execution activity.
type ExecutionResult struct {
	WorkerID  string      `json:"worker_id"`
	Type      string      `json:"type,omitempty"`
	Processed int         `json:"processed"`
	Completed int         `json:"completed"`
	Failed    int         `json:"failed"`
	LastJob   *domain.Job `json:"last_job,omitempty"`
}

// BackgroundRunConfig controls the optional worker background drain loop.
type BackgroundRunConfig struct {
	WorkerID string
	Type     string
	Limit    int
	Interval time.Duration
}

const (
	jobQueued    = "queued"
	jobClaimed   = "claimed"
	jobCompleted = "completed"
	jobFailed    = "failed"
)

// WorkerService provides an in-memory async job queue prototype.
type WorkerService struct {
	mu       sync.RWMutex
	jobs     map[string]domain.Job
	order    []string
	handlers map[string]JobHandler
	now      func() time.Time
	newJobID func() (string, error)
}

// NewWorkerService constructs an in-memory worker service.
func NewWorkerService() *WorkerService {
	return &WorkerService{
		jobs:     make(map[string]domain.Job),
		order:    make([]string, 0),
		handlers: make(map[string]JobHandler),
		now:      time.Now,
		newJobID: func() (string, error) {
			return idgen.Token(8)
		},
	}
}

// RegisterHandler binds a job type to a worker execution handler.
func (s *WorkerService) RegisterHandler(jobType string, handler JobHandler) {
	if jobType == "" || handler == nil {
		return
	}

	s.mu.Lock()
	s.handlers[jobType] = handler
	s.mu.Unlock()
}

// Enqueue creates a queued async job.
func (s *WorkerService) Enqueue(jobType string, payload string) (domain.Job, *apperrors.Error) {
	if jobType == "" {
		err := apperrors.New("invalid_request", "type is required", http.StatusBadRequest)
		return domain.Job{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	jobID, idErr := s.newJobID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.Job{}, &internal
	}

	job := domain.Job{
		ID:        jobID,
		Type:      jobType,
		Payload:   payload,
		Status:    jobQueued,
		Attempts:  0,
		CreatedAt: s.now(),
	}

	s.jobs[job.ID] = job
	s.order = append(s.order, job.ID)
	return job, nil
}

// ClaimNext returns the oldest queued or failed job matching the optional type filter.
func (s *WorkerService) ClaimNext(workerID string, jobType string) (domain.Job, *apperrors.Error) {
	if workerID == "" {
		err := apperrors.New("invalid_request", "worker_id is required", http.StatusBadRequest)
		return domain.Job{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for _, id := range s.order {
		job := s.jobs[id]
		if jobType != "" && job.Type != jobType {
			continue
		}
		if job.Status != jobQueued && job.Status != jobFailed {
			continue
		}

		job.Status = jobClaimed
		job.Attempts++
		job.ClaimedBy = workerID
		job.ClaimedAt = &now
		job.LastError = ""
		s.jobs[id] = job
		return job, nil
	}

	err := apperrors.New("not_found", "no claimable job found", http.StatusNotFound)
	return domain.Job{}, &err
}

// Complete marks a claimed job as completed.
func (s *WorkerService) Complete(jobID string, workerID string) (domain.Job, *apperrors.Error) {
	return s.transition(jobID, workerID, jobCompleted, "")
}

// Fail marks a claimed job as failed so it can be retried later.
func (s *WorkerService) Fail(jobID string, workerID string, lastError string) (domain.Job, *apperrors.Error) {
	if lastError == "" {
		err := apperrors.New("invalid_request", "last_error is required", http.StatusBadRequest)
		return domain.Job{}, &err
	}
	return s.transition(jobID, workerID, jobFailed, lastError)
}

// ListJobs returns jobs filtered by optional status and type.
func (s *WorkerService) ListJobs(status string, jobType string) ([]domain.Job, *apperrors.Error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]domain.Job, 0, len(s.order))
	for _, id := range s.order {
		job := s.jobs[id]
		if status != "" && job.Status != status {
			continue
		}
		if jobType != "" && job.Type != jobType {
			continue
		}
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

// ExecuteNext claims and executes the next available job for the worker.
func (s *WorkerService) ExecuteNext(ctx context.Context, workerID string, jobType string) (ExecutionResult, *apperrors.Error) {
	job, appErr := s.ClaimNext(workerID, jobType)
	if appErr != nil {
		return ExecutionResult{}, appErr
	}

	result := ExecutionResult{
		WorkerID:  workerID,
		Type:      jobType,
		Processed: 1,
		LastJob:   &job,
	}

	handler := s.lookupHandler(job.Type)
	if handler == nil {
		lastError := fmt.Sprintf("no handler registered for job type %s", job.Type)
		failed, failErr := s.Fail(job.ID, workerID, lastError)
		if failErr != nil {
			return ExecutionResult{}, failErr
		}
		result.Failed = 1
		result.LastJob = &failed
		return result, nil
	}

	if handlerErr := handler(ctx, job); handlerErr != nil {
		failed, failErr := s.Fail(job.ID, workerID, handlerErr.Message)
		if failErr != nil {
			return ExecutionResult{}, failErr
		}
		result.Failed = 1
		result.LastJob = &failed
		return result, nil
	}

	completed, completeErr := s.Complete(job.ID, workerID)
	if completeErr != nil {
		return ExecutionResult{}, completeErr
	}
	result.Completed = 1
	result.LastJob = &completed
	return result, nil
}

// ExecuteUntilEmpty drains claimable jobs for the optional type filter.
func (s *WorkerService) ExecuteUntilEmpty(ctx context.Context, workerID string, jobType string, limit int) (ExecutionResult, *apperrors.Error) {
	if workerID == "" {
		err := apperrors.New("invalid_request", "worker_id is required", http.StatusBadRequest)
		return ExecutionResult{}, &err
	}
	if limit <= 0 {
		limit = 100
	}

	result := ExecutionResult{
		WorkerID: workerID,
		Type:     jobType,
	}

	for i := 0; i < limit; i++ {
		next, appErr := s.ExecuteNext(ctx, workerID, jobType)
		if appErr != nil {
			if appErr.Code == "not_found" {
				return result, nil
			}
			return ExecutionResult{}, appErr
		}

		result.Processed += next.Processed
		result.Completed += next.Completed
		result.Failed += next.Failed
		result.LastJob = next.LastJob
	}

	return result, nil
}

// RunBackground drains jobs on a ticker until the context is cancelled.
func (s *WorkerService) RunBackground(ctx context.Context, cfg BackgroundRunConfig) *apperrors.Error {
	if cfg.WorkerID == "" {
		err := apperrors.New("invalid_request", "worker_id is required", http.StatusBadRequest)
		return &err
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 250 * time.Millisecond
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 100
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		if _, appErr := s.ExecuteUntilEmpty(ctx, cfg.WorkerID, cfg.Type, cfg.Limit); appErr != nil {
			return appErr
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (s *WorkerService) transition(jobID string, workerID string, targetStatus string, lastError string) (domain.Job, *apperrors.Error) {
	if jobID == "" || workerID == "" {
		err := apperrors.New("invalid_request", "job_id and worker_id are required", http.StatusBadRequest)
		return domain.Job{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		err := apperrors.New("not_found", "job not found", http.StatusNotFound)
		return domain.Job{}, &err
	}

	if job.Status != jobClaimed {
		err := apperrors.New("invalid_state", "job must be claimed before transition", http.StatusConflict)
		return domain.Job{}, &err
	}

	if job.ClaimedBy != workerID {
		err := apperrors.New("forbidden", "job is claimed by another worker", http.StatusForbidden)
		return domain.Job{}, &err
	}

	now := s.now()
	job.Status = targetStatus
	job.LastError = lastError
	if targetStatus == jobCompleted {
		job.CompletedAt = &now
	}
	s.jobs[jobID] = job
	return job, nil
}

func (s *WorkerService) lookupHandler(jobType string) JobHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.handlers[jobType]
}

func (s *WorkerService) String() string {
	return fmt.Sprintf("worker-service(jobs=%d)", len(s.jobs))
}
