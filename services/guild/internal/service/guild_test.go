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

type fakePresenceReader struct {
	snapshots map[string]PresenceSnapshot
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

func (f *fakePresenceReader) GetPresence(_ context.Context, playerID string) (PresenceSnapshot, *apperrors.Error) {
	snapshot, ok := f.snapshots[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", 404)
		return PresenceSnapshot{}, &err
	}
	return snapshot, nil
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

	svc := NewGuildService(invites, nil)
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

func TestListMemberStatesIncludesPresence(t *testing.T) {
	t.Parallel()

	svc := NewGuildService(&fakeInviteClient{}, &fakePresenceReader{
		snapshots: map[string]PresenceSnapshot{
			"p1": {
				PlayerID:  "p1",
				Status:    presenceOnline,
				SessionID: "sess-1",
			},
		},
	})
	svc.newGuildID = func() (string, error) { return "guild-1", nil }

	guild, err := svc.CreateGuild("Guild", "p1")
	if err != nil {
		t.Fatalf("create guild returned error: %+v", err)
	}

	states, listErr := svc.ListMemberStates(context.Background(), guild.ID)
	if listErr != nil {
		t.Fatalf("list member states returned error: %+v", listErr)
	}

	if len(states) != 1 || states[0].Presence != presenceOnline {
		t.Fatalf("unexpected member states: %+v", states)
	}
}

func TestGuildServiceWithInjectedStore(t *testing.T) {
	t.Parallel()

	guilds := newMemoryGuildStore()
	svc := NewGuildServiceWithStore(guilds, &fakeInviteClient{}, nil)
	svc.newGuildID = func() (string, error) { return "guild-1", nil }

	guild, err := svc.CreateGuild("Guild", "p1")
	if err != nil {
		t.Fatalf("create guild returned error: %+v", err)
	}

	stored, ok, getErr := guilds.GetGuild(guild.ID)
	if getErr != nil {
		t.Fatalf("guild store get returned error: %v", getErr)
	}
	if !ok || stored.ID != guild.ID {
		t.Fatalf("unexpected stored guild: %+v", stored)
	}
}

func TestTransferOwnershipAndKickMember(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{
		get: Invite{
			ID:         "inv-1",
			Domain:     inviteDomainGuild,
			ResourceID: "guild-1",
			ToPlayerID: "p2",
			Status:     "accepted",
		},
	}

	svc := NewGuildService(invites, nil)
	svc.newGuildID = func() (string, error) { return "guild-1", nil }

	guild, err := svc.CreateGuild("Guild", "p1")
	if err != nil {
		t.Fatalf("create guild returned error: %+v", err)
	}
	if _, joinErr := svc.JoinWithInvite(context.Background(), guild.ID, "inv-1", "p2"); joinErr != nil {
		t.Fatalf("join returned error: %+v", joinErr)
	}

	transferred, transferErr := svc.TransferOwnership(guild.ID, "p1", "p2")
	if transferErr != nil {
		t.Fatalf("transfer ownership returned error: %+v", transferErr)
	}
	if transferred.OwnerID != "p2" {
		t.Fatalf("unexpected owner after transfer: %+v", transferred)
	}

	kicked, kickErr := svc.KickMember(guild.ID, "p2", "p1")
	if kickErr != nil {
		t.Fatalf("kick member returned error: %+v", kickErr)
	}
	if len(kicked.Members) != 1 || kicked.Members[0].PlayerID != "p2" || kicked.Members[0].Role != roleOwner {
		t.Fatalf("unexpected guild after kick: %+v", kicked)
	}
}
