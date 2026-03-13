package service

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

func TestEnqueueClaimAndCompleteJob(t *testing.T) {
	t.Parallel()

	svc := NewWorkerService()
	job, err := svc.Enqueue("invite.expire", `{"invite_id":"inv-1"}`)
	if err != nil {
		t.Fatalf("enqueue returned error: %+v", err)
	}

	claimed, claimErr := svc.ClaimNext("worker-a", "invite.expire")
	if claimErr != nil {
		t.Fatalf("claim returned error: %+v", claimErr)
	}

	if claimed.ID != job.ID || claimed.Status != jobClaimed {
		t.Fatalf("unexpected claimed job: %+v", claimed)
	}

	completed, completeErr := svc.Complete(job.ID, "worker-a")
	if completeErr != nil {
		t.Fatalf("complete returned error: %+v", completeErr)
	}

	if completed.Status != jobCompleted {
		t.Fatalf("unexpected completed status: %+v", completed)
	}
}

func TestFailMakesJobRetryable(t *testing.T) {
	t.Parallel()

	svc := NewWorkerService()
	job, err := svc.Enqueue("chat.replay_backfill", `{}`)
	if err != nil {
		t.Fatalf("enqueue returned error: %+v", err)
	}

	if _, claimErr := svc.ClaimNext("worker-a", ""); claimErr != nil {
		t.Fatalf("claim returned error: %+v", claimErr)
	}

	failed, failErr := svc.Fail(job.ID, "worker-a", "temporary failure")
	if failErr != nil {
		t.Fatalf("fail returned error: %+v", failErr)
	}

	if failed.Status != jobFailed {
		t.Fatalf("unexpected failed status: %+v", failed)
	}

	retry, retryErr := svc.ClaimNext("worker-b", "chat.replay_backfill")
	if retryErr != nil {
		t.Fatalf("retry claim returned error: %+v", retryErr)
	}

	if retry.Attempts != 2 {
		t.Fatalf("unexpected attempt count after retry: %+v", retry)
	}
}

func TestExecuteNextCompletesRegisteredHandler(t *testing.T) {
	t.Parallel()

	svc := NewWorkerService()
	svc.RegisterHandler("invite.expire", func(context.Context, domain.Job) *apperrors.Error {
		return nil
	})

	if _, err := svc.Enqueue("invite.expire", `{}`); err != nil {
		t.Fatalf("enqueue returned error: %+v", err)
	}

	result, err := svc.ExecuteNext(context.Background(), "worker-a", "invite.expire")
	if err != nil {
		t.Fatalf("execute next returned error: %+v", err)
	}
	if result.Completed != 1 || result.LastJob == nil || result.LastJob.Status != jobCompleted {
		t.Fatalf("unexpected execution result: %+v", result)
	}
}

func TestExecuteNextFailsHandlerError(t *testing.T) {
	t.Parallel()

	svc := NewWorkerService()
	svc.RegisterHandler("chat.offline_delivery", func(context.Context, domain.Job) *apperrors.Error {
		err := apperrors.New("delivery_failed", "temporary failure", 500)
		return &err
	})

	if _, err := svc.Enqueue("chat.offline_delivery", `{}`); err != nil {
		t.Fatalf("enqueue returned error: %+v", err)
	}

	result, err := svc.ExecuteNext(context.Background(), "worker-a", "chat.offline_delivery")
	if err != nil {
		t.Fatalf("execute next returned error: %+v", err)
	}
	if result.Failed != 1 || result.LastJob == nil || result.LastJob.Status != jobFailed {
		t.Fatalf("unexpected execution result: %+v", result)
	}
}
