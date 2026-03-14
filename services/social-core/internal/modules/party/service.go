package party

import (
	"net/http"
	"slices"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	invitemodule "github.com/xyun1996/social_backend/services/social-core/internal/modules/invite"
)

type Party struct {
	ID        string    `json:"id"`
	LeaderID  string    `json:"leader_id"`
	MemberIDs []string  `json:"member_ids"`
	CreatedAt time.Time `json:"created_at"`
}

type ReadyState struct {
	PartyID   string    `json:"party_id"`
	PlayerID  string    `json:"player_id"`
	IsReady   bool      `json:"is_ready"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MemberState struct {
	PlayerID string `json:"player_id"`
	IsLeader bool   `json:"is_leader"`
	IsReady  bool   `json:"is_ready"`
}

type InviteBoundary interface {
	CreateInvite(domainName, resourceID, fromPlayerID, toPlayerID string, ttl time.Duration) (invitemodule.Invite, *apperrors.Error)
	GetInvite(inviteID string) (invitemodule.Invite, *apperrors.Error)
}

type Service struct {
	mu      sync.RWMutex
	now     func() time.Time
	invites InviteBoundary
	parties map[string]Party
	ready   map[string]map[string]ReadyState
}

func NewService(invites InviteBoundary) *Service {
	return &Service{
		now:     time.Now,
		invites: invites,
		parties: make(map[string]Party),
		ready:   make(map[string]map[string]ReadyState),
	}
}

func (s *Service) CreateParty(leaderID string) (Party, *apperrors.Error) {
	if leaderID == "" {
		err := apperrors.New("invalid_request", "leader_id is required", http.StatusBadRequest)
		return Party{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, party := range s.parties {
		if slices.Contains(party.MemberIDs, leaderID) {
			err := apperrors.New("already_in_party", "leader already belongs to a party", http.StatusConflict)
			return Party{}, &err
		}
	}

	partyID, err := idgen.Token(8)
	if err != nil {
		internal := apperrors.Internal()
		return Party{}, &internal
	}
	record := Party{
		ID:        partyID,
		LeaderID:  leaderID,
		MemberIDs: []string{leaderID},
		CreatedAt: s.now(),
	}
	s.parties[record.ID] = record
	return record, nil
}

func (s *Service) GetParty(partyID string) (Party, *apperrors.Error) {
	if partyID == "" {
		err := apperrors.New("invalid_request", "party_id is required", http.StatusBadRequest)
		return Party{}, &err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.parties[partyID]
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return Party{}, &err
	}
	return record, nil
}

func (s *Service) FindPartyByPlayer(playerID string) (Party, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return Party{}, &err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, party := range s.parties {
		if slices.Contains(party.MemberIDs, playerID) {
			return party, nil
		}
	}
	err := apperrors.New("not_found", "party not found for player", http.StatusNotFound)
	return Party{}, &err
}

func (s *Service) CreateInvite(partyID, actorPlayerID, toPlayerID string) (invitemodule.Invite, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id, actor_player_id, and to_player_id are required", http.StatusBadRequest)
		return invitemodule.Invite{}, &err
	}
	s.mu.RLock()
	record, ok := s.parties[partyID]
	s.mu.RUnlock()
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return invitemodule.Invite{}, &err
	}
	if record.LeaderID != actorPlayerID {
		err := apperrors.New("forbidden", "only the party leader can invite", http.StatusForbidden)
		return invitemodule.Invite{}, &err
	}
	if slices.Contains(record.MemberIDs, toPlayerID) {
		err := apperrors.New("already_member", "player is already in the party", http.StatusConflict)
		return invitemodule.Invite{}, &err
	}
	return s.invites.CreateInvite("party", partyID, actorPlayerID, toPlayerID, 0)
}

func (s *Service) JoinWithInvite(partyID, inviteID, actorPlayerID string) (Party, *apperrors.Error) {
	if partyID == "" || inviteID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id, invite_id, and actor_player_id are required", http.StatusBadRequest)
		return Party{}, &err
	}
	invite, appErr := s.invites.GetInvite(inviteID)
	if appErr != nil {
		return Party{}, appErr
	}
	if invite.Domain != "party" || invite.ResourceID != partyID {
		err := apperrors.New("forbidden", "invite does not belong to this party", http.StatusForbidden)
		return Party{}, &err
	}
	if invite.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "invite belongs to a different player", http.StatusForbidden)
		return Party{}, &err
	}
	if invite.Status != invitemodule.StatusAccepted {
		err := apperrors.New("invite_not_accepted", "invite must be accepted before joining", http.StatusConflict)
		return Party{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.parties[partyID]
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return Party{}, &err
	}
	if slices.Contains(record.MemberIDs, actorPlayerID) {
		return record, nil
	}
	record.MemberIDs = append(record.MemberIDs, actorPlayerID)
	slices.Sort(record.MemberIDs)
	s.parties[record.ID] = record
	return record, nil
}

func (s *Service) SetReady(partyID, actorPlayerID string, isReady bool) (ReadyState, *apperrors.Error) {
	if partyID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "party_id and actor_player_id are required", http.StatusBadRequest)
		return ReadyState{}, &err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.parties[partyID]
	if !ok {
		err := apperrors.New("not_found", "party not found", http.StatusNotFound)
		return ReadyState{}, &err
	}
	if !slices.Contains(record.MemberIDs, actorPlayerID) {
		err := apperrors.New("forbidden", "only party members can update ready state", http.StatusForbidden)
		return ReadyState{}, &err
	}
	if _, ok := s.ready[partyID]; !ok {
		s.ready[partyID] = make(map[string]ReadyState)
	}
	state := ReadyState{
		PartyID:   partyID,
		PlayerID:  actorPlayerID,
		IsReady:   isReady,
		UpdatedAt: s.now(),
	}
	s.ready[partyID][actorPlayerID] = state
	return state, nil
}

func (s *Service) ListReadyStates(partyID string) ([]ReadyState, *apperrors.Error) {
	record, appErr := s.GetParty(partyID)
	if appErr != nil {
		return nil, appErr
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	states := make([]ReadyState, 0, len(record.MemberIDs))
	for _, memberID := range record.MemberIDs {
		state, ok := s.ready[partyID][memberID]
		if !ok {
			state = ReadyState{PartyID: partyID, PlayerID: memberID, IsReady: false}
		}
		states = append(states, state)
	}
	return states, nil
}

func (s *Service) ListMemberStates(partyID string) ([]MemberState, *apperrors.Error) {
	record, appErr := s.GetParty(partyID)
	if appErr != nil {
		return nil, appErr
	}
	readyStates, appErr := s.ListReadyStates(partyID)
	if appErr != nil {
		return nil, appErr
	}
	readyByPlayer := make(map[string]ReadyState, len(readyStates))
	for _, state := range readyStates {
		readyByPlayer[state.PlayerID] = state
	}
	states := make([]MemberState, 0, len(record.MemberIDs))
	for _, memberID := range record.MemberIDs {
		states = append(states, MemberState{
			PlayerID: memberID,
			IsLeader: memberID == record.LeaderID,
			IsReady:  readyByPlayer[memberID].IsReady,
		})
	}
	return states, nil
}
