package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/party/internal/domain"
)

const inviteDomainParty = "party"

const (
	presenceOnline  = "online"
	presenceOffline = "offline"
)

// Invite contains the subset of invite state party depends on.
type Invite struct {
	ID           string `json:"id"`
	Domain       string `json:"domain"`
	ResourceID   string `json:"resource_id,omitempty"`
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
	Status       string `json:"status"`
}

// InviteClient is the explicit boundary from party to invite.
type InviteClient interface {
	CreateInvite(ctx context.Context, domainName string, resourceID string, fromPlayerID string, toPlayerID string) (Invite, *apperrors.Error)
	GetInvite(ctx context.Context, inviteID string) (Invite, *apperrors.Error)
}

// PresenceSnapshot contains the subset of presence state party uses.
type PresenceSnapshot struct {
	PlayerID  string `json:"player_id"`
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// PresenceReader resolves current presence state for party members.
type PresenceReader interface {
	GetPresence(ctx context.Context, playerID string) (PresenceSnapshot, *apperrors.Error)
}

// MemberState combines membership, ready state, and presence state.
type MemberState struct {
	PlayerID  string `json:"player_id"`
	IsLeader  bool   `json:"is_leader"`
	IsReady   bool   `json:"is_ready"`
	Presence  string `json:"presence"`
	SessionID string `json:"session_id,omitempty"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// PartyService provides an in-memory prototype for leader, member, and ready flows.
type PartyService struct {
	parties    PartyStore
	ready      ReadyStateStore
	invites    InviteClient
	presence   PresenceReader
	now        func() time.Time
	newPartyID func() (string, error)
}

// NewPartyService constructs an in-memory party service.
func NewPartyService(invites InviteClient, presence PresenceReader) *PartyService {
	return NewPartyServiceWithStores(newMemoryPartyStore(), newMemoryReadyStateStore(), invites, presence)
}

// NewPartyServiceWithStores constructs the party service with injected persistence boundaries.
func NewPartyServiceWithStores(parties PartyStore, ready ReadyStateStore, invites InviteClient, presence PresenceReader) *PartyService {
	return &PartyService{
		parties:  parties,
		ready:    ready,
		invites:  invites,
		presence: presence,
		now:      time.Now,
		newPartyID: func() (string, error) {
			return idgen.Token(8)
		},
	}
}

// CreateParty creates a new party with a single leader member.
func (s *PartyService) CreateParty(leaderID string) (domain.Party, *apperrors.Error) {
	if leaderID == "" {
		err := apperrors.New("invalid_request", "leader_id is required", http.StatusBadRequest)
		return domain.Party{}, &err
	}

	partyID, idErr := s.newPartyID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}

	party := domain.Party{
		ID:        partyID,
		LeaderID:  leaderID,
		MemberIDs: []string{leaderID},
		CreatedAt: s.now(),
	}

	if err := s.parties.SaveParty(party); err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	return party, nil
}

// GetParty returns a party by id.
func (s *PartyService) GetParty(partyID string) (domain.Party, *apperrors.Error) {
	if partyID == "" {
		err := apperrors.New("invalid_request", "party_id is required", http.StatusBadRequest)
		return domain.Party{}, &err
	}

	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.Party{}, &err
	}

	return party, nil
}

// CreateInvite delegates invite lifecycle creation to the invite service.
func (s *PartyService) CreateInvite(ctx context.Context, partyID string, actorPlayerID string, toPlayerID string) (Invite, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id, actor_player_id, and to_player_id are required", http.StatusBadRequest)
		return Invite{}, &err
	}

	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return Invite{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return Invite{}, &err
	}

	if party.LeaderID != actorPlayerID {
		err := apperrors.New("forbidden", "only the party leader can invite", http.StatusForbidden)
		return Invite{}, &err
	}

	if slices.Contains(party.MemberIDs, toPlayerID) {
		err := apperrors.New("already_member", "player is already in the party", http.StatusConflict)
		return Invite{}, &err
	}

	return s.invites.CreateInvite(ctx, inviteDomainParty, partyID, actorPlayerID, toPlayerID)
}

// JoinWithInvite adds the invited player after invite acceptance.
func (s *PartyService) JoinWithInvite(ctx context.Context, partyID string, inviteID string, actorPlayerID string) (domain.Party, *apperrors.Error) {
	if partyID == "" || inviteID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id, invite_id, and actor_player_id are required", http.StatusBadRequest)
		return domain.Party{}, &err
	}

	invite, appErr := s.invites.GetInvite(ctx, inviteID)
	if appErr != nil {
		return domain.Party{}, appErr
	}

	if invite.Domain != inviteDomainParty || invite.ResourceID != partyID {
		err := apperrors.New("forbidden", "invite does not belong to this party", http.StatusForbidden)
		return domain.Party{}, &err
	}

	if invite.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "invite belongs to a different player", http.StatusForbidden)
		return domain.Party{}, &err
	}

	if invite.Status != "accepted" {
		err := apperrors.New("invite_not_accepted", "invite must be accepted before joining", http.StatusConflict)
		return domain.Party{}, &err
	}

	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.Party{}, &err
	}

	if slices.Contains(party.MemberIDs, actorPlayerID) {
		return party, nil
	}

	party.MemberIDs = append(party.MemberIDs, actorPlayerID)
	slices.Sort(party.MemberIDs)
	if err := s.parties.SaveParty(party); err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	return party, nil
}

// SetReady updates a member's ready state.
func (s *PartyService) SetReady(partyID string, actorPlayerID string, isReady bool) (domain.ReadyState, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id and actor_player_id are required", http.StatusBadRequest)
		return domain.ReadyState{}, &err
	}

	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.ReadyState{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.ReadyState{}, &err
	}

	if !slices.Contains(party.MemberIDs, actorPlayerID) {
		err := apperrors.New("forbidden", "only party members can update ready state", http.StatusForbidden)
		return domain.ReadyState{}, &err
	}

	if s.presence != nil {
		snapshot, appErr := s.presence.GetPresence(context.Background(), actorPlayerID)
		if appErr != nil && appErr.Code != "not_found" {
			return domain.ReadyState{}, appErr
		}
		if appErr != nil || snapshot.Status != presenceOnline {
			err := apperrors.New("presence_required", "player must be online to update ready state", http.StatusConflict)
			return domain.ReadyState{}, &err
		}
	}

	state := domain.ReadyState{
		PartyID:   partyID,
		PlayerID:  actorPlayerID,
		IsReady:   isReady,
		UpdatedAt: s.now(),
	}

	if err := s.ready.SaveReadyState(state); err != nil {
		internal := apperrors.Internal()
		return domain.ReadyState{}, &internal
	}
	return state, nil
}

// ListReadyStates returns ready status for current party members.
func (s *PartyService) ListReadyStates(partyID string) ([]domain.ReadyState, *apperrors.Error) {
	if partyID == "" {
		err := apperrors.New("invalid_request", "party_id is required", http.StatusBadRequest)
		return nil, &err
	}

	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return nil, &err
	}

	storedStates, err := s.ready.ListReadyStates(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	byPlayer := make(map[string]domain.ReadyState, len(storedStates))
	for _, state := range storedStates {
		byPlayer[state.PlayerID] = state
	}

	readyStates := make([]domain.ReadyState, 0, len(party.MemberIDs))
	for _, memberID := range party.MemberIDs {
		state, ok := byPlayer[memberID]
		if !ok {
			state = domain.ReadyState{
				PartyID:  partyID,
				PlayerID: memberID,
				IsReady:  false,
			}
		}
		readyStates = append(readyStates, state)
	}

	return readyStates, nil
}

// ListMemberStates returns ready and presence state for current members.
func (s *PartyService) ListMemberStates(ctx context.Context, partyID string) ([]MemberState, *apperrors.Error) {
	if partyID == "" {
		err := apperrors.New("invalid_request", "party_id is required", http.StatusBadRequest)
		return nil, &err
	}

	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return nil, &err
	}

	readyStates, err := s.ready.ListReadyStates(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	readyByPlayer := make(map[string]domain.ReadyState, len(readyStates))
	for _, state := range readyStates {
		readyByPlayer[state.PlayerID] = state
	}

	states := make([]MemberState, 0, len(party.MemberIDs))
	for _, memberID := range party.MemberIDs {
		memberState := MemberState{
			PlayerID: memberID,
			IsLeader: memberID == party.LeaderID,
			Presence: presenceOffline,
		}

		if ready, ok := readyByPlayer[memberID]; ok {
			memberState.IsReady = ready.IsReady
		}

		if s.presence != nil {
			snapshot, appErr := s.presence.GetPresence(ctx, memberID)
			if appErr != nil && appErr.Code != "not_found" {
				return nil, appErr
			}
			if appErr == nil {
				memberState.Presence = snapshot.Status
				memberState.SessionID = snapshot.SessionID
				memberState.RealmID = snapshot.RealmID
				memberState.Location = snapshot.Location
			}
		}

		states = append(states, memberState)
	}

	return states, nil
}

func (s *PartyService) String() string {
	parties, err := s.parties.ListParties()
	if err != nil {
		return "party-service(parties=unknown)"
	}
	return fmt.Sprintf("party-service(parties=%d)", len(parties))
}
