package service

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

type fakeInviteClient struct {
	create Invite
	get    Invite
}

func (f *fakeInviteClient) CreateInvite(_ context.Context, domainName string, resourceID string, fromPlayerID string, toPlayerID string) (Invite, *apperrors.Error) {
	result := f.create
	result.Domain = domainName
	result.ResourceID = resourceID
	result.FromPlayerID = fromPlayerID
	result.ToPlayerID = toPlayerID
	return result, nil
}

func (f *fakeInviteClient) GetInvite(_ context.Context, _ string) (Invite, *apperrors.Error) {
	return f.get, nil
}

func TestCreateInviteAndJoinParty(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{
		create: Invite{ID: "inv-1", Status: "pending"},
		get: Invite{
			ID:           "inv-1",
			Domain:       inviteDomainParty,
			ResourceID:   "party-1",
			FromPlayerID: "p1",
			ToPlayerID:   "p2",
			Status:       "accepted",
		},
	}

	svc := NewPartyService(invites)
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}

	invite, inviteErr := svc.CreateInvite(context.Background(), party.ID, "p1", "p2")
	if inviteErr != nil {
		t.Fatalf("create invite returned error: %+v", inviteErr)
	}

	if invite.ResourceID != party.ID {
		t.Fatalf("unexpected invite resource id: %+v", invite)
	}

	joined, joinErr := svc.JoinWithInvite(context.Background(), party.ID, invite.ID, "p2")
	if joinErr != nil {
		t.Fatalf("join returned error: %+v", joinErr)
	}

	if len(joined.MemberIDs) != 2 {
		t.Fatalf("unexpected member count: %+v", joined.MemberIDs)
	}
}

func TestReadyStateRequiresMembership(t *testing.T) {
	t.Parallel()

	svc := NewPartyService(&fakeInviteClient{})
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}

	if _, readyErr := svc.SetReady(party.ID, "p2", true); readyErr == nil {
		t.Fatalf("expected non-member ready update to fail")
	}
}
