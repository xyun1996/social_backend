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
	presenceOnline      = "online"
	presenceOffline     = "offline"
	queueStatusOpen     = "queued"
	queueStatusAssigned = "assigned"
	queueStatusLeft     = "left"
	queueStatusResolved = "resolved"
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
	queues     QueueStateStore
	invites    InviteClient
	presence   PresenceReader
	now        func() time.Time
	newPartyID func() (string, error)
	queueTTL   time.Duration
}

// NewPartyService constructs an in-memory party service.
func NewPartyService(invites InviteClient, presence PresenceReader) *PartyService {
	return NewPartyServiceWithStores(newMemoryPartyStore(), newMemoryReadyStateStore(), newMemoryQueueStateStore(), invites, presence)
}

// NewPartyServiceWithStores constructs the party service with injected persistence boundaries.
func NewPartyServiceWithStores(parties PartyStore, ready ReadyStateStore, queues QueueStateStore, invites InviteClient, presence PresenceReader) *PartyService {
	return &PartyService{
		parties:  parties,
		ready:    ready,
		queues:   queues,
		invites:  invites,
		presence: presence,
		now:      time.Now,
		newPartyID: func() (string, error) {
			return idgen.Token(8)
		},
		queueTTL: 10 * time.Minute,
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
	if appErr := s.ensureQueueUnlocked(partyID); appErr != nil {
		return Invite{}, appErr
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
	if appErr := s.ensureQueueUnlocked(partyID); appErr != nil {
		return domain.Party{}, appErr
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
	if appErr := s.ensureQueueUnlocked(partyID); appErr != nil {
		return domain.ReadyState{}, appErr
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

// JoinQueue enrolls a fully online and ready party into a named social queue.
func (s *PartyService) JoinQueue(ctx context.Context, partyID string, actorPlayerID string, queueName string) (domain.QueueState, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" || queueName == "" {
		err := apperrors.New("invalid_request", "party_id, actor_player_id, and queue_name are required", http.StatusBadRequest)
		return domain.QueueState{}, &err
	}

	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueState{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.QueueState{}, &err
	}
	if party.LeaderID != actorPlayerID {
		err := apperrors.New("forbidden", "only the party leader can join queue", http.StatusForbidden)
		return domain.QueueState{}, &err
	}

	current, exists, err := s.getActiveQueueState(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueState{}, &internal
	}
	if exists {
		if current.QueueName == queueName && current.Status == queueStatusOpen {
			return current, nil
		}
		if current.Status == queueStatusAssigned {
			err := apperrors.New("match_assigned", "party already has an assigned match", http.StatusConflict)
			return domain.QueueState{}, &err
		}
		err := apperrors.New("already_queued", "party is already queued in a different queue", http.StatusConflict)
		return domain.QueueState{}, &err
	}

	if appErr := s.requireQueueReady(ctx, party); appErr != nil {
		return domain.QueueState{}, appErr
	}

	joinedAt := s.now()
	expiresAt := joinedAt.Add(s.queueTTL)
	state := domain.QueueState{
		PartyID: partyID,
		QueueName: queueName,
		Status: queueStatusOpen,
		JoinedBy: actorPlayerID,
		JoinedAt: joinedAt,
		ExpiresAt: &expiresAt,
	}
	if err := s.queues.SaveQueueState(state); err != nil {
		internal := apperrors.Internal()
		return domain.QueueState{}, &internal
	}
	if err := s.queues.DeleteQueueAssignment(partyID); err != nil {
		internal := apperrors.Internal()
		return domain.QueueState{}, &internal
	}
	return state, nil
}

// GetQueueState returns the active queue enrollment for a party.
func (s *PartyService) GetQueueState(partyID string) (domain.QueueState, *apperrors.Error) {
	if partyID == "" {
		err := apperrors.New("invalid_request", "party_id is required", http.StatusBadRequest)
		return domain.QueueState{}, &err
	}

	if _, ok, appErr := s.getParty(partyID); appErr != nil {
		return domain.QueueState{}, appErr
	} else if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.QueueState{}, &err
	}

	state, ok, err := s.getActiveQueueState(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueState{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party queue state not found", http.StatusNotFound)
		return domain.QueueState{}, &err
	}
	return state, nil
}

// GetQueueHandoff returns the external queue handoff snapshot for the active queued party.
func (s *PartyService) GetQueueHandoff(ctx context.Context, partyID string) (domain.QueueHandoff, []MemberState, *apperrors.Error) {
	if partyID == "" {
		err := apperrors.New("invalid_request", "party_id is required", http.StatusBadRequest)
		return domain.QueueHandoff{}, nil, &err
	}

	party, ok, appErr := s.getParty(partyID)
	if appErr != nil {
		return domain.QueueHandoff{}, nil, appErr
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.QueueHandoff{}, nil, &err
	}

	state, ok, err := s.getActiveQueueState(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueHandoff{}, nil, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party queue state not found", http.StatusNotFound)
		return domain.QueueHandoff{}, nil, &err
	}
	if state.Status != queueStatusOpen {
		err := apperrors.New("match_assigned", "party queue handoff is no longer open", http.StatusConflict)
		return domain.QueueHandoff{}, nil, &err
	}

	members, appErr := s.ListMemberStates(ctx, partyID)
	if appErr != nil {
		return domain.QueueHandoff{}, nil, appErr
	}

	handoff := domain.QueueHandoff{
		TicketID:  queueTicketID(party.ID, state.QueueName, state.JoinedAt),
		PartyID:   party.ID,
		QueueName: state.QueueName,
		LeaderID:  party.LeaderID,
		MemberIDs: append([]string(nil), party.MemberIDs...),
		JoinedAt:  state.JoinedAt,
	}
	return handoff, members, nil
}

// AssignMatch records the callback payload after an external matchmaker consumes a queue handoff.
func (s *PartyService) AssignMatch(ctx context.Context, partyID string, ticketID string, matchID string, serverID string, connectionHint string) (domain.QueueAssignment, *apperrors.Error) {
	if partyID == "" || ticketID == "" || matchID == "" {
		err := apperrors.New("invalid_request", "party_id, ticket_id, and match_id are required", http.StatusBadRequest)
		return domain.QueueAssignment{}, &err
	}

	party, ok, appErr := s.getParty(partyID)
	if appErr != nil {
		return domain.QueueAssignment{}, appErr
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.QueueAssignment{}, &err
	}

	state, ok, err := s.getActiveQueueState(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueAssignment{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party queue state not found", http.StatusNotFound)
		return domain.QueueAssignment{}, &err
	}

	expectedTicketID := queueTicketID(party.ID, state.QueueName, state.JoinedAt)
	if ticketID != expectedTicketID {
		err := apperrors.New("ticket_mismatch", "ticket_id does not match the active queue handoff", http.StatusConflict)
		return domain.QueueAssignment{}, &err
	}

	if state.Status == queueStatusAssigned {
		assignment, ok, err := s.queues.GetQueueAssignment(partyID)
		if err != nil {
			internal := apperrors.Internal()
			return domain.QueueAssignment{}, &internal
		}
		if ok && assignment.TicketID == ticketID && assignment.MatchID == matchID {
			return assignment, nil
		}
		conflict := apperrors.New("match_assigned", "party already has an assigned match", http.StatusConflict)
		return domain.QueueAssignment{}, &conflict
	}

	if appErr := s.requireQueueReady(ctx, party); appErr != nil {
		return domain.QueueAssignment{}, appErr
	}

	assignment := domain.QueueAssignment{
		TicketID:       ticketID,
		PartyID:        partyID,
		QueueName:      state.QueueName,
		MatchID:        matchID,
		Status:         queueStatusAssigned,
		ServerID:       serverID,
		ConnectionHint: connectionHint,
		AssignedAt:     s.now(),
	}
	state.Status = queueStatusAssigned

	if err := s.queues.SaveQueueState(state); err != nil {
		internal := apperrors.Internal()
		return domain.QueueAssignment{}, &internal
	}
	if err := s.queues.SaveQueueAssignment(assignment); err != nil {
		internal := apperrors.Internal()
		return domain.QueueAssignment{}, &internal
	}

	return assignment, nil
}

// GetQueueAssignment returns the current match assignment for a queued party.
func (s *PartyService) GetQueueAssignment(partyID string) (domain.QueueAssignment, *apperrors.Error) {
	if partyID == "" {
		err := apperrors.New("invalid_request", "party_id is required", http.StatusBadRequest)
		return domain.QueueAssignment{}, &err
	}

	if _, ok, appErr := s.getParty(partyID); appErr != nil {
		return domain.QueueAssignment{}, appErr
	} else if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.QueueAssignment{}, &err
	}

	assignment, ok, err := s.queues.GetQueueAssignment(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueAssignment{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party queue assignment not found", http.StatusNotFound)
		return domain.QueueAssignment{}, &err
	}
	return assignment, nil
}

// ResolveMatch clears the assigned queue ownership after a match is consumed or cancelled.
func (s *PartyService) ResolveMatch(partyID string, ticketID string, matchID string, status string) (domain.QueueResolution, *apperrors.Error) {
	if partyID == "" || ticketID == "" || matchID == "" || status == "" {
		err := apperrors.New("invalid_request", "party_id, ticket_id, match_id, and status are required", http.StatusBadRequest)
		return domain.QueueResolution{}, &err
	}

	if _, ok, appErr := s.getParty(partyID); appErr != nil {
		return domain.QueueResolution{}, appErr
	} else if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.QueueResolution{}, &err
	}

	assignment, ok, err := s.queues.GetQueueAssignment(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueResolution{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party queue assignment not found", http.StatusNotFound)
		return domain.QueueResolution{}, &err
	}
	if assignment.TicketID != ticketID || assignment.MatchID != matchID {
		err := apperrors.New("assignment_mismatch", "match resolution does not match the active queue assignment", http.StatusConflict)
		return domain.QueueResolution{}, &err
	}

	if err := s.queues.DeleteQueueAssignment(partyID); err != nil {
		internal := apperrors.Internal()
		return domain.QueueResolution{}, &internal
	}
	if err := s.queues.DeleteQueueState(partyID); err != nil {
		internal := apperrors.Internal()
		return domain.QueueResolution{}, &internal
	}

	return domain.QueueResolution{
		TicketID:   ticketID,
		PartyID:    partyID,
		QueueName:  assignment.QueueName,
		MatchID:    matchID,
		Status:     status,
		ResolvedAt: s.now(),
	}, nil
}

// FindPartyByPlayer returns the current party membership snapshot for a player.
func (s *PartyService) FindPartyByPlayer(ctx context.Context, playerID string) (domain.Party, []MemberState, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return domain.Party{}, nil, &err
	}

	parties, err := s.parties.ListParties()
	if err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, nil, &internal
	}
	for _, party := range parties {
		if !slices.Contains(party.MemberIDs, playerID) {
			continue
		}
		members, appErr := s.ListMemberStates(ctx, party.ID)
		if appErr != nil {
			return domain.Party{}, nil, appErr
		}
		return party, members, nil
	}

	notFound := apperrors.New("not_found", "party not found for player", http.StatusNotFound)
	return domain.Party{}, nil, &notFound
}

// LeaveQueue removes the active queue enrollment for a party.
func (s *PartyService) LeaveQueue(partyID string, actorPlayerID string) (domain.QueueLeaveResult, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id and actor_player_id are required", http.StatusBadRequest)
		return domain.QueueLeaveResult{}, &err
	}

	party, ok, appErr := s.getParty(partyID)
	if appErr != nil {
		return domain.QueueLeaveResult{}, appErr
	}
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return domain.QueueLeaveResult{}, &err
	}
	if party.LeaderID != actorPlayerID {
		err := apperrors.New("forbidden", "only the party leader can leave queue", http.StatusForbidden)
		return domain.QueueLeaveResult{}, &err
	}

	state, ok, err := s.getActiveQueueState(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.QueueLeaveResult{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "party queue state not found", http.StatusNotFound)
		return domain.QueueLeaveResult{}, &err
	}
	if state.Status == queueStatusAssigned {
		err := apperrors.New("match_assigned", "assigned matches cannot leave queue through the social queue endpoint", http.StatusConflict)
		return domain.QueueLeaveResult{}, &err
	}
	if err := s.queues.DeleteQueueState(partyID); err != nil {
		internal := apperrors.Internal()
		return domain.QueueLeaveResult{}, &internal
	}
	if err := s.queues.DeleteQueueAssignment(partyID); err != nil {
		internal := apperrors.Internal()
		return domain.QueueLeaveResult{}, &internal
	}

	return domain.QueueLeaveResult{
		PartyID:   partyID,
		QueueName: state.QueueName,
		Status:    queueStatusLeft,
		LeftAt:    s.now(),
	}, nil
}

// LeaveParty removes a non-leader member from the party and clears their ready state.
func (s *PartyService) LeaveParty(partyID string, actorPlayerID string) (domain.Party, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id and actor_player_id are required", http.StatusBadRequest)
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
	if !slices.Contains(party.MemberIDs, actorPlayerID) {
		err := apperrors.New("forbidden", "only party members can leave", http.StatusForbidden)
		return domain.Party{}, &err
	}
	if appErr := s.ensureQueueUnlocked(partyID); appErr != nil {
		return domain.Party{}, appErr
	}
	if party.LeaderID == actorPlayerID {
		err := apperrors.New("leader_must_transfer", "party leader must transfer leadership before leaving", http.StatusConflict)
		return domain.Party{}, &err
	}

	party.MemberIDs = deleteMemberIDs(party.MemberIDs, actorPlayerID)
	if err := s.parties.SaveParty(party); err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	if err := s.ready.DeleteReadyState(partyID, actorPlayerID); err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	return party, nil
}

// KickMember removes a member at the direction of the party leader.
func (s *PartyService) KickMember(partyID string, actorPlayerID string, targetPlayerID string) (domain.Party, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" || targetPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id, actor_player_id, and target_player_id are required", http.StatusBadRequest)
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
	if party.LeaderID != actorPlayerID {
		err := apperrors.New("forbidden", "only the party leader can kick members", http.StatusForbidden)
		return domain.Party{}, &err
	}
	if appErr := s.ensureQueueUnlocked(partyID); appErr != nil {
		return domain.Party{}, appErr
	}
	if targetPlayerID == party.LeaderID {
		err := apperrors.New("invalid_request", "party leader cannot kick themselves", http.StatusBadRequest)
		return domain.Party{}, &err
	}
	if !slices.Contains(party.MemberIDs, targetPlayerID) {
		err := apperrors.New("not_found", "target member not found", http.StatusNotFound)
		return domain.Party{}, &err
	}

	party.MemberIDs = deleteMemberIDs(party.MemberIDs, targetPlayerID)
	if err := s.parties.SaveParty(party); err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	if err := s.ready.DeleteReadyState(partyID, targetPlayerID); err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	return party, nil
}

// TransferLeader assigns party leadership to another current member.
func (s *PartyService) TransferLeader(partyID string, actorPlayerID string, targetPlayerID string) (domain.Party, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" || targetPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id, actor_player_id, and target_player_id are required", http.StatusBadRequest)
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
	if party.LeaderID != actorPlayerID {
		err := apperrors.New("forbidden", "only the party leader can transfer leadership", http.StatusForbidden)
		return domain.Party{}, &err
	}
	if appErr := s.ensureQueueUnlocked(partyID); appErr != nil {
		return domain.Party{}, appErr
	}
	if !slices.Contains(party.MemberIDs, targetPlayerID) {
		err := apperrors.New("not_found", "target member not found", http.StatusNotFound)
		return domain.Party{}, &err
	}

	party.LeaderID = targetPlayerID
	if err := s.parties.SaveParty(party); err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, &internal
	}
	return party, nil
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

func deleteMemberIDs(memberIDs []string, targetPlayerID string) []string {
	filtered := memberIDs[:0]
	for _, memberID := range memberIDs {
		if memberID == targetPlayerID {
			continue
		}
		filtered = append(filtered, memberID)
	}
	return filtered
}

func (s *PartyService) getParty(partyID string) (domain.Party, bool, *apperrors.Error) {
	party, ok, err := s.parties.GetParty(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Party{}, false, &internal
	}
	return party, ok, nil
}



func queueTicketID(partyID string, queueName string, joinedAt time.Time) string {
	return fmt.Sprintf("ticket:%s:%s:%d", partyID, queueName, joinedAt.UTC().Unix())
}

func (s *PartyService) requireQueueReady(ctx context.Context, party domain.Party) *apperrors.Error {
	readyStates, err := s.ready.ListReadyStates(party.ID)
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}

	readyByPlayer := make(map[string]domain.ReadyState, len(readyStates))
	for _, state := range readyStates {
		readyByPlayer[state.PlayerID] = state
	}

	for _, memberID := range party.MemberIDs {
		ready, ok := readyByPlayer[memberID]
		if !ok || !ready.IsReady {
			err := apperrors.New("party_not_ready", "all party members must be ready before joining queue", http.StatusConflict)
			return &err
		}
		if s.presence != nil {
			snapshot, appErr := s.presence.GetPresence(ctx, memberID)
			if appErr != nil || snapshot.Status != presenceOnline {
				err := apperrors.New("presence_required", "all party members must be online before joining queue", http.StatusConflict)
				return &err
			}
		}
	}

	return nil
}





