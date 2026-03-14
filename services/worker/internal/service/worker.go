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

type JobHandler func(ctx context.Context, job domain.Job) *apperrors.Error

type ExecutionResult struct {
	WorkerID  string      `json:"worker_id"`
	Type      string      `json:"type,omitempty"`
	Processed int         `json:"processed"`
	Completed int         `json:"completed"`
	Failed    int         `json:"failed"`
	Dead      int         `json:"dead"`
	LastJob   *domain.Job `json:"last_job,omitempty"`
}

type BackgroundRunConfig struct {
	WorkerID string
	Type     string
	Limit    int
	Interval time.Duration
}

type EnqueueOptions struct { MaxAttempts int }

const (
	jobQueued     = "queued"
	jobClaimed    = "claimed"
	jobCompleted  = "completed"
	jobFailed     = "failed"
	jobDeadLetter = "dead_letter"
)

type WorkerService struct {
	store    JobStore
	mu       sync.RWMutex
	handlers map[string]JobHandler
	now      func() time.Time
	newJobID func() (string, error)
}

func NewWorkerService() *WorkerService { return NewWorkerServiceWithStore(newMemoryJobStore()) }
func NewWorkerServiceWithStore(store JobStore) *WorkerService {
	return &WorkerService{store: store, handlers: make(map[string]JobHandler), now: time.Now, newJobID: func() (string, error) { return idgen.Token(8) }}
}
func (s *WorkerService) RegisterHandler(jobType string, handler JobHandler) { if jobType == "" || handler == nil { return }; s.mu.Lock(); s.handlers[jobType] = handler; s.mu.Unlock() }
func (s *WorkerService) Enqueue(jobType string, payload string) (domain.Job, *apperrors.Error) { return s.EnqueueWithOptions(jobType, payload, EnqueueOptions{}) }

func (s *WorkerService) EnqueueWithOptions(jobType string, payload string, options EnqueueOptions) (domain.Job, *apperrors.Error) {
	if jobType == "" { err := apperrors.New("invalid_request", "type is required", http.StatusBadRequest); return domain.Job{}, &err }
	jobID, idErr := s.newJobID(); if idErr != nil { internal := apperrors.Internal(); return domain.Job{}, &internal }
	if options.MaxAttempts <= 0 { options.MaxAttempts = 3 }
	job := domain.Job{ID: jobID, Type: jobType, Payload: payload, Status: jobQueued, Attempts: 0, MaxAttempts: options.MaxAttempts, CreatedAt: s.now()}
	if err := s.store.SaveJob(job); err != nil { internal := apperrors.Internal(); return domain.Job{}, &internal }
	return job, nil
}

func (s *WorkerService) ClaimNext(workerID string, jobType string) (domain.Job, *apperrors.Error) {
	if workerID == "" { err := apperrors.New("invalid_request", "worker_id is required", http.StatusBadRequest); return domain.Job{}, &err }
	now := s.now(); jobs, err := s.store.ListJobs(); if err != nil { internal := apperrors.Internal(); return domain.Job{}, &internal }
	for _, job := range jobs {
		if jobType != "" && job.Type != jobType { continue }
		if job.Status != jobQueued && job.Status != jobFailed { continue }
		if job.NextAttemptAt != nil && job.NextAttemptAt.After(now) { continue }
		job.Status = jobClaimed
		job.Attempts++
		job.ClaimedBy = workerID
		job.ClaimedAt = &now
		job.CompletedAt = nil
		job.LastError = ""
		job.NextAttemptAt = nil
		if err := s.store.SaveJob(job); err != nil { internal := apperrors.Internal(); return domain.Job{}, &internal }
		return job, nil
	}
	notFound := apperrors.New("not_found", "no claimable job found", http.StatusNotFound)
	return domain.Job{}, &notFound
}

func (s *WorkerService) Complete(jobID string, workerID string) (domain.Job, *apperrors.Error) { return s.transition(jobID, workerID, jobCompleted, "") }
func (s *WorkerService) Fail(jobID string, workerID string, lastError string) (domain.Job, *apperrors.Error) {
	if lastError == "" { err := apperrors.New("invalid_request", "last_error is required", http.StatusBadRequest); return domain.Job{}, &err }
	return s.transition(jobID, workerID, jobFailed, lastError)
}

func (s *WorkerService) ListJobs(status string, jobType string) ([]domain.Job, *apperrors.Error) {
	jobs, err := s.store.ListJobs(); if err != nil { internal := apperrors.Internal(); return nil, &internal }
	filtered := make([]domain.Job, 0, len(jobs))
	for _, job := range jobs { if status != "" && job.Status != status { continue }; if jobType != "" && job.Type != jobType { continue }; filtered = append(filtered, job) }
	slices.SortFunc(filtered, func(a, b domain.Job) int {
		if !a.CreatedAt.Equal(b.CreatedAt) { if a.CreatedAt.Before(b.CreatedAt) { return -1 }; return 1 }
		switch { case a.ID < b.ID: return -1; case a.ID > b.ID: return 1; default: return 0 }
	})
	return filtered, nil
}

func (s *WorkerService) ExecuteNext(ctx context.Context, workerID string, jobType string) (ExecutionResult, *apperrors.Error) {
	job, appErr := s.ClaimNext(workerID, jobType); if appErr != nil { return ExecutionResult{}, appErr }
	result := ExecutionResult{WorkerID: workerID, Type: jobType, Processed: 1, LastJob: &job}
	handler := s.lookupHandler(job.Type)
	if handler == nil {
		lastError := fmt.Sprintf("no handler registered for job type %s", job.Type)
		failed, failErr := s.Fail(job.ID, workerID, lastError); if failErr != nil { return ExecutionResult{}, failErr }
		if failed.Status == jobDeadLetter { result.Dead = 1 } else { result.Failed = 1 }
		result.LastJob = &failed
		return result, nil
	}
	if handlerErr := handler(ctx, job); handlerErr != nil {
		failed, failErr := s.Fail(job.ID, workerID, handlerErr.Message); if failErr != nil { return ExecutionResult{}, failErr }
		if failed.Status == jobDeadLetter { result.Dead = 1 } else { result.Failed = 1 }
		result.LastJob = &failed
		return result, nil
	}
	completed, completeErr := s.Complete(job.ID, workerID); if completeErr != nil { return ExecutionResult{}, completeErr }
	result.Completed = 1; result.LastJob = &completed; return result, nil
}

func (s *WorkerService) ExecuteUntilEmpty(ctx context.Context, workerID string, jobType string, limit int) (ExecutionResult, *apperrors.Error) {
	if workerID == "" { err := apperrors.New("invalid_request", "worker_id is required", http.StatusBadRequest); return ExecutionResult{}, &err }
	if limit <= 0 { limit = 100 }
	result := ExecutionResult{WorkerID: workerID, Type: jobType}
	for i := 0; i < limit; i++ {
		next, appErr := s.ExecuteNext(ctx, workerID, jobType)
		if appErr != nil {
			if appErr.Code == "not_found" { return result, nil }
			return ExecutionResult{}, appErr
		}
		result.Processed += next.Processed; result.Completed += next.Completed; result.Failed += next.Failed; result.Dead += next.Dead; result.LastJob = next.LastJob
	}
	return result, nil
}

func (s *WorkerService) RunBackground(ctx context.Context, cfg BackgroundRunConfig) *apperrors.Error {
	if cfg.WorkerID == "" { err := apperrors.New("invalid_request", "worker_id is required", http.StatusBadRequest); return &err }
	if cfg.Interval <= 0 { cfg.Interval = 250 * time.Millisecond }
	if cfg.Limit <= 0 { cfg.Limit = 100 }
	ticker := time.NewTicker(cfg.Interval); defer ticker.Stop()
	for {
		if _, appErr := s.ExecuteUntilEmpty(ctx, cfg.WorkerID, cfg.Type, cfg.Limit); appErr != nil { return appErr }
		select { case <-ctx.Done(): return nil; case <-ticker.C: }
	}
}

func (s *WorkerService) transition(jobID string, workerID string, targetStatus string, lastError string) (domain.Job, *apperrors.Error) {
	if jobID == "" || workerID == "" { err := apperrors.New("invalid_request", "job_id and worker_id are required", http.StatusBadRequest); return domain.Job{}, &err }
	job, ok, err := s.store.GetJob(jobID); if err != nil { internal := apperrors.Internal(); return domain.Job{}, &internal }
	if !ok { err := apperrors.New("not_found", "job not found", http.StatusNotFound); return domain.Job{}, &err }
	if job.Status != jobClaimed { err := apperrors.New("invalid_state", "job must be claimed before transition", http.StatusConflict); return domain.Job{}, &err }
	if job.ClaimedBy != workerID { err := apperrors.New("forbidden", "job is claimed by another worker", http.StatusForbidden); return domain.Job{}, &err }
	now := s.now(); job.LastError = lastError
	switch targetStatus {
	case jobCompleted:
		job.Status = jobCompleted
		job.CompletedAt = &now
		job.NextAttemptAt = nil
	case jobFailed:
		if job.MaxAttempts > 0 && job.Attempts >= job.MaxAttempts {
			job.Status = jobDeadLetter
			job.CompletedAt = &now
			job.NextAttemptAt = nil
		} else {
			job.Status = jobFailed
			job.CompletedAt = nil
			next := now.Add(backoffForAttempt(job.Attempts))
			job.NextAttemptAt = &next
		}
	default:
		job.Status = targetStatus
	}
	if err := s.store.SaveJob(job); err != nil { internal := apperrors.Internal(); return domain.Job{}, &internal }
	return job, nil
}

func backoffForAttempt(attempt int) time.Duration { if attempt <= 1 { return 0 }; return time.Duration((attempt-1)*(attempt-1)) * time.Second }
func (s *WorkerService) lookupHandler(jobType string) JobHandler { s.mu.RLock(); defer s.mu.RUnlock(); return s.handlers[jobType] }
func (s *WorkerService) String() string { jobs, err := s.store.ListJobs(); if err != nil { return "worker-service(jobs=unknown)" }; return fmt.Sprintf("worker-service(jobs=%d)", len(jobs)) }

