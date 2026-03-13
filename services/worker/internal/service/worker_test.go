package service

import "testing"

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
