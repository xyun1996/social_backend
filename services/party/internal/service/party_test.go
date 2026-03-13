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

	svc := NewPartyService(invites, nil)
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

	svc := NewPartyService(&fakeInviteClient{}, nil)
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}

	if _, readyErr := svc.SetReady(party.ID, "p2", true); readyErr == nil {
		t.Fatalf("expected non-member ready update to fail")
	}
}

func TestReadyRequiresOnlinePresence(t *testing.T) {
	t.Parallel()

	svc := NewPartyService(&fakeInviteClient{}, &fakePresenceReader{
		snapshots: map[string]PresenceSnapshot{
			"p1": {
				PlayerID: "p1",
				Status:   presenceOffline,
			},
		},
	})
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}

	if _, readyErr := svc.SetReady(party.ID, "p1", true); readyErr == nil {
		t.Fatalf("expected offline member ready update to fail")
	}
}

func TestListMemberStatesIncludesPresence(t *testing.T) {
	t.Parallel()

	svc := NewPartyService(&fakeInviteClient{}, &fakePresenceReader{
		snapshots: map[string]PresenceSnapshot{
			"p1": {
				PlayerID:  "p1",
				Status:    presenceOnline,
				SessionID: "sess-1",
			},
		},
	})
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}

	states, listErr := svc.ListMemberStates(context.Background(), party.ID)
	if listErr != nil {
		t.Fatalf("list member states returned error: %+v", listErr)
	}

	if len(states) != 1 || states[0].Presence != presenceOnline {
		t.Fatalf("unexpected member states: %+v", states)
	}
}

func TestPartyServiceWithInjectedStores(t *testing.T) {
	t.Parallel()

	parties := newMemoryPartyStore()
	ready := newMemoryReadyStateStore()
	svc := NewPartyServiceWithStores(parties, ready, &fakeInviteClient{}, nil)
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}

	if _, readyErr := svc.SetReady(party.ID, "p1", true); readyErr != nil {
		t.Fatalf("set ready returned error: %+v", readyErr)
	}

	storedParty, ok, getErr := parties.GetParty(party.ID)
	if getErr != nil {
		t.Fatalf("party store get returned error: %v", getErr)
	}
	if !ok || storedParty.ID != party.ID {
		t.Fatalf("unexpected stored party: %+v", storedParty)
	}

	readyStates, listErr := ready.ListReadyStates(party.ID)
	if listErr != nil {
		t.Fatalf("ready store list returned error: %v", listErr)
	}
	if len(readyStates) != 1 || !readyStates[0].IsReady {
		t.Fatalf("unexpected stored ready states: %+v", readyStates)
	}
}

func TestTransferLeaderAndLeaveParty(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{
		get: Invite{
			ID:         "inv-1",
			Domain:     inviteDomainParty,
			ResourceID: "party-1",
			ToPlayerID: "p2",
			Status:     "accepted",
		},
	}

	svc := NewPartyService(invites, nil)
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}
	if _, joinErr := svc.JoinWithInvite(context.Background(), party.ID, "inv-1", "p2"); joinErr != nil {
		t.Fatalf("join returned error: %+v", joinErr)
	}

	transferred, transferErr := svc.TransferLeader(party.ID, "p1", "p2")
	if transferErr != nil {
		t.Fatalf("transfer returned error: %+v", transferErr)
	}
	if transferred.LeaderID != "p2" {
		t.Fatalf("unexpected leader after transfer: %+v", transferred)
	}

	left, leaveErr := svc.LeaveParty(party.ID, "p1")
	if leaveErr != nil {
		t.Fatalf("leave returned error: %+v", leaveErr)
	}
	if len(left.MemberIDs) != 1 || left.MemberIDs[0] != "p2" {
		t.Fatalf("unexpected party after leave: %+v", left)
	}
}

func TestKickMemberClearsReadyState(t *testing.T) {
	t.Parallel()

	invites := &fakeInviteClient{
		get: Invite{
			ID:         "inv-1",
			Domain:     inviteDomainParty,
			ResourceID: "party-1",
			ToPlayerID: "p2",
			Status:     "accepted",
		},
	}
	presence := &fakePresenceReader{
		snapshots: map[string]PresenceSnapshot{
			"p2": {PlayerID: "p2", Status: presenceOnline},
		},
	}

	svc := NewPartyService(invites, presence)
	svc.newPartyID = func() (string, error) { return "party-1", nil }

	party, err := svc.CreateParty("p1")
	if err != nil {
		t.Fatalf("create party returned error: %+v", err)
	}
	if _, joinErr := svc.JoinWithInvite(context.Background(), party.ID, "inv-1", "p2"); joinErr != nil {
		t.Fatalf("join returned error: %+v", joinErr)
	}
	if _, readyErr := svc.SetReady(party.ID, "p2", true); readyErr != nil {
		t.Fatalf("set ready returned error: %+v", readyErr)
	}

	kicked, kickErr := svc.KickMember(party.ID, "p1", "p2")
	if kickErr != nil {
		t.Fatalf("kick returned error: %+v", kickErr)
	}
	if len(kicked.MemberIDs) != 1 || kicked.MemberIDs[0] != "p1" {
		t.Fatalf("unexpected party after kick: %+v", kicked)
	}

	readyStates, listErr := svc.ListReadyStates(party.ID)
	if listErr != nil {
		t.Fatalf("list ready states returned error: %+v", listErr)
	}
	if len(readyStates) != 1 || readyStates[0].PlayerID != "p1" {
		t.Fatalf("unexpected ready states after kick: %+v", readyStates)
	}
}
