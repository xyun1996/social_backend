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

func TestCreateInviteAndJoinGuild(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{
		create: Invite{ID: "inv-1", Status: "pending"},
		get: Invite{
			ID:           "inv-1",
			Domain:       inviteDomainGuild,
			ResourceID:   "guild-1",
			FromPlayerID: "p1",
			ToPlayerID:   "p2",
			Status:       "accepted",
		},
	}

	svc := NewGuildService(invites)
	svc.newGuildID = func() (string, error) { return "guild-1", nil }

	guild, err := svc.CreateGuild("Test Guild", "p1")
	if err != nil {
		t.Fatalf("create guild returned error: %+v", err)
	}

	invite, inviteErr := svc.CreateInvite(context.Background(), guild.ID, "p1", "p2")
	if inviteErr != nil {
		t.Fatalf("create invite returned error: %+v", inviteErr)
	}

	if invite.ResourceID != guild.ID {
		t.Fatalf("unexpected invite resource id: %+v", invite)
	}

	joined, joinErr := svc.JoinWithInvite(context.Background(), guild.ID, invite.ID, "p2")
	if joinErr != nil {
		t.Fatalf("join returned error: %+v", joinErr)
	}

	if len(joined.Members) != 2 {
		t.Fatalf("unexpected guild member count: %+v", joined.Members)
	}
}
