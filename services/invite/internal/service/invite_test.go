package service

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/invite/internal/domain"
)

type fakeScheduler struct {
	jobType string
	payload string
	calls   int
	err     *apperrors.Error
}

func (f *fakeScheduler) EnqueueJob(_ context.Context, jobType string, payload string) *apperrors.Error {
	f.jobType = jobType
	f.payload = payload
	f.calls++
	return f.err
}

type recordingInviteStore struct {
	invites map[string]domain.Invite
}

func newRecordingInviteStore() *recordingInviteStore {
	return &recordingInviteStore{invites: make(map[string]domain.Invite)}
}

func (s *recordingInviteStore) ListInvites() ([]domain.Invite, error) {
	invites := make([]domain.Invite, 0, len(s.invites))
	for _, invite := range s.invites {
		invites = append(invites, invite)
	}
	return invites, nil
}

func (s *recordingInviteStore) SaveInvite(invite domain.Invite) error {
	s.invites[invite.ID] = invite
	return nil
}

func (s *recordingInviteStore) GetInvite(inviteID string) (domain.Invite, bool, error) {
	invite, ok := s.invites[inviteID]
	return invite, ok, nil
}

func TestCreateAndAcceptInvite(t *testing.T) {
	t.Parallel()

	svc := NewInviteService(nil)
	invite, err := svc.CreateInvite("party", "party-1", "p1", "p2", time.Minute)
	if err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}

	if invite.Status != inviteStatusPending {
		t.Fatalf("unexpected initial status: %q", invite.Status)
	}

	accepted, respondErr := svc.RespondInvite(invite.ID, "p2", inviteActionAccept)
	if respondErr != nil {
		t.Fatalf("accept returned error: %+v", respondErr)
	}

	if accepted.Status != inviteStatusAccepted {
		t.Fatalf("unexpected accepted status: %q", accepted.Status)
	}

	if accepted.RespondedAt == nil {
		t.Fatalf("expected responded_at to be set")
	}
}

func TestExpiredInviteCannotBeAccepted(t *testing.T) {
	t.Parallel()

	svc := NewInviteService(nil)
	base := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return base }

	invite, err := svc.CreateInvite("guild", "guild-1", "p1", "p2", time.Second)
	if err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}

	svc.now = func() time.Time { return base.Add(2 * time.Second) }

	expired, respondErr := svc.RespondInvite(invite.ID, "p2", inviteActionAccept)
	if respondErr == nil {
		t.Fatalf("expected expired invite response to fail")
	}

	if expired.Status != inviteStatusExpired {
		t.Fatalf("unexpected expired status: %q", expired.Status)
	}
}

func TestListInvitesFiltersInboxPending(t *testing.T) {
	t.Parallel()

	svc := NewInviteService(nil)
	if _, err := svc.CreateInvite("party", "party-1", "p1", "p2", time.Minute); err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}

	accepted, err := svc.CreateInvite("guild", "guild-1", "p3", "p2", time.Minute)
	if err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}

	if _, respondErr := svc.RespondInvite(accepted.ID, "p2", inviteActionAccept); respondErr != nil {
		t.Fatalf("accept returned error: %+v", respondErr)
	}

	inbox, listErr := svc.ListInvites("p2", "inbox", inviteStatusPending)
	if listErr != nil {
		t.Fatalf("list invites returned error: %+v", listErr)
	}

	if len(inbox) != 1 {
		t.Fatalf("unexpected inbox size: %d", len(inbox))
	}

	if inbox[0].FromPlayerID != "p1" {
		t.Fatalf("unexpected invite in inbox: %+v", inbox[0])
	}
}

func TestCreateInviteEnqueuesExpiryJob(t *testing.T) {
	t.Parallel()

	scheduler := &fakeScheduler{}
	svc := NewInviteService(scheduler)

	invite, err := svc.CreateInvite("party", "party-1", "p1", "p2", time.Minute)
	if err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}

	if scheduler.calls != 1 {
		t.Fatalf("expected one enqueue call, got %d", scheduler.calls)
	}

	if scheduler.jobType != expireJobType {
		t.Fatalf("unexpected job type: %q", scheduler.jobType)
	}

	if scheduler.payload == "" {
		t.Fatalf("expected enqueue payload to be populated")
	}

	if invite.ID == "" {
		t.Fatalf("expected invite id to be populated")
	}
}

func TestDuplicatePendingInviteDoesNotEnqueueAgain(t *testing.T) {
	t.Parallel()

	scheduler := &fakeScheduler{}
	svc := NewInviteService(scheduler)

	first, err := svc.CreateInvite("party", "party-1", "p1", "p2", time.Minute)
	if err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}

	second, err := svc.CreateInvite("party", "party-1", "p1", "p2", time.Minute)
	if err != nil {
		t.Fatalf("second create invite returned error: %+v", err)
	}

	if scheduler.calls != 1 {
		t.Fatalf("expected one enqueue call, got %d", scheduler.calls)
	}

	if first.ID != second.ID {
		t.Fatalf("expected duplicate pending invite to return same id: %q vs %q", first.ID, second.ID)
	}
}

func TestExpireInviteTransitionsPendingInvite(t *testing.T) {
	t.Parallel()

	svc := NewInviteService(nil)
	invite, err := svc.CreateInvite("party", "party-1", "p1", "p2", time.Minute)
	if err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}

	expired, err := svc.ExpireInvite(invite.ID)
	if err != nil {
		t.Fatalf("expire invite returned error: %+v", err)
	}
	if expired.Status != inviteStatusExpired {
		t.Fatalf("unexpected expired status: %q", expired.Status)
	}
}

func TestInviteServiceUsesInjectedStore(t *testing.T) {
	t.Parallel()

	store := newRecordingInviteStore()
	svc := NewInviteServiceWithStore(store, nil)

	invite, err := svc.CreateInvite("party", "party-1", "p1", "p2", time.Minute)
	if err != nil {
		t.Fatalf("create invite returned error: %+v", err)
	}
	responded, err := svc.RespondInvite(invite.ID, "p2", inviteActionAccept)
	if err != nil {
		t.Fatalf("respond invite returned error: %+v", err)
	}

	if stored, ok := store.invites[invite.ID]; !ok {
		t.Fatalf("expected invite to be stored")
	} else if stored.Status != responded.Status {
		t.Fatalf("unexpected stored invite status: %+v", stored)
	}
}
