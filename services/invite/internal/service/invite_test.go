package service

import (
	"testing"
	"time"
)

func TestCreateAndAcceptInvite(t *testing.T) {
	t.Parallel()

	svc := NewInviteService()
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

	svc := NewInviteService()
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

	svc := NewInviteService()
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
