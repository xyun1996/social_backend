package invite

import (
	"net/http"
	"slices"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
)

const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusDeclined = "declined"
	StatusExpired  = "expired"
	StatusCanceled = "canceled"

	ActionAccept  = "accept"
	ActionDecline = "decline"

	DefaultTTL = 15 * time.Minute
)

type Invite struct {
	ID           string     `json:"id"`
	Domain       string     `json:"domain"`
	ResourceID   string     `json:"resource_id,omitempty"`
	FromPlayerID string     `json:"from_player_id"`
	ToPlayerID   string     `json:"to_player_id"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	RespondedAt  *time.Time `json:"responded_at,omitempty"`
}

type Service struct {
	mu      sync.RWMutex
	now     func() time.Time
	invites map[string]Invite
}

func NewService() *Service {
	return &Service{
		now:     time.Now,
		invites: make(map[string]Invite),
	}
}

func (s *Service) CreateInvite(domainName, resourceID, fromPlayerID, toPlayerID string, ttl time.Duration) (Invite, *apperrors.Error) {
	if domainName == "" || fromPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "domain, from_player_id, and to_player_id are required", http.StatusBadRequest)
		return Invite{}, &err
	}
	if fromPlayerID == toPlayerID {
		err := apperrors.New("invalid_request", "cannot invite yourself", http.StatusBadRequest)
		return Invite{}, &err
	}
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	for id, invite := range s.invites {
		if invite.Status == StatusPending && s.isExpired(invite, now) {
			invite.Status = StatusExpired
			s.invites[id] = invite
		}
		if invite.Domain == domainName &&
			invite.ResourceID == resourceID &&
			invite.FromPlayerID == fromPlayerID &&
			invite.ToPlayerID == toPlayerID &&
			invite.Status == StatusPending {
			return invite, nil
		}
	}

	inviteID, err := idgen.Token(8)
	if err != nil {
		internal := apperrors.Internal()
		return Invite{}, &internal
	}
	record := Invite{
		ID:           inviteID,
		Domain:       domainName,
		ResourceID:   resourceID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		Status:       StatusPending,
		CreatedAt:    now,
		ExpiresAt:    now.Add(ttl),
	}
	s.invites[record.ID] = record
	return record, nil
}

func (s *Service) GetInvite(inviteID string) (Invite, *apperrors.Error) {
	if inviteID == "" {
		err := apperrors.New("invalid_request", "invite_id is required", http.StatusBadRequest)
		return Invite{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.invites[inviteID]
	if !ok {
		err := apperrors.New("not_found", "invite not found", http.StatusNotFound)
		return Invite{}, &err
	}
	if record.Status == StatusPending && s.isExpired(record, s.now()) {
		record.Status = StatusExpired
		s.invites[inviteID] = record
	}
	return record, nil
}

func (s *Service) RespondInvite(inviteID, actorPlayerID, action string) (Invite, *apperrors.Error) {
	if inviteID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "invite_id and actor_player_id are required", http.StatusBadRequest)
		return Invite{}, &err
	}
	if action != ActionAccept && action != ActionDecline {
		err := apperrors.New("invalid_request", "action must be accept or decline", http.StatusBadRequest)
		return Invite{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.invites[inviteID]
	if !ok {
		err := apperrors.New("not_found", "invite not found", http.StatusNotFound)
		return Invite{}, &err
	}
	if record.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the invited player can respond", http.StatusForbidden)
		return Invite{}, &err
	}
	if s.isExpired(record, s.now()) {
		record.Status = StatusExpired
		s.invites[inviteID] = record
		err := apperrors.New("invite_expired", "invite has expired", http.StatusConflict)
		return record, &err
	}
	switch record.Status {
	case StatusAccepted, StatusDeclined:
		return record, nil
	case StatusExpired:
		err := apperrors.New("invite_expired", "invite has expired", http.StatusConflict)
		return record, &err
	case StatusCanceled:
		err := apperrors.New("invite_not_pending", "only pending invites can be updated", http.StatusConflict)
		return record, &err
	}

	now := s.now()
	record.RespondedAt = &now
	if action == ActionAccept {
		record.Status = StatusAccepted
	} else {
		record.Status = StatusDeclined
	}
	s.invites[inviteID] = record
	return record, nil
}

func (s *Service) CancelInvite(inviteID, actorPlayerID string) (Invite, *apperrors.Error) {
	if inviteID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "invite_id and actor_player_id are required", http.StatusBadRequest)
		return Invite{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.invites[inviteID]
	if !ok {
		err := apperrors.New("not_found", "invite not found", http.StatusNotFound)
		return Invite{}, &err
	}
	if record.FromPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the invite sender can cancel", http.StatusForbidden)
		return Invite{}, &err
	}
	switch record.Status {
	case StatusAccepted, StatusDeclined, StatusExpired:
		err := apperrors.New("invite_not_pending", "only pending invites can be canceled", http.StatusConflict)
		return record, &err
	case StatusCanceled:
		return record, nil
	}
	now := s.now()
	record.Status = StatusCanceled
	record.RespondedAt = &now
	s.invites[inviteID] = record
	return record, nil
}

func (s *Service) ExpireInvite(inviteID string) (Invite, *apperrors.Error) {
	if inviteID == "" {
		err := apperrors.New("invalid_request", "invite_id is required", http.StatusBadRequest)
		return Invite{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.invites[inviteID]
	if !ok {
		err := apperrors.New("not_found", "invite not found", http.StatusNotFound)
		return Invite{}, &err
	}
	if record.Status == StatusPending {
		record.Status = StatusExpired
		s.invites[inviteID] = record
	}
	return record, nil
}

func (s *Service) ListInvites(playerID, role, status string) ([]Invite, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}
	if role == "" {
		role = "all"
	}
	if role != "all" && role != "inbox" && role != "outbox" {
		err := apperrors.New("invalid_request", "role must be all, inbox, or outbox", http.StatusBadRequest)
		return nil, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	list := make([]Invite, 0)
	for id, invite := range s.invites {
		if invite.Status == StatusPending && s.isExpired(invite, now) {
			invite.Status = StatusExpired
			s.invites[id] = invite
		}
		if !matchesRole(invite, playerID, role) {
			continue
		}
		if status != "" && invite.Status != status {
			continue
		}
		list = append(list, invite)
	}
	slices.SortFunc(list, func(a, b Invite) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return list, nil
}

func (s *Service) isExpired(invite Invite, now time.Time) bool {
	return !invite.ExpiresAt.After(now)
}

func matchesRole(invite Invite, playerID, role string) bool {
	switch role {
	case "inbox":
		return invite.ToPlayerID == playerID
	case "outbox":
		return invite.FromPlayerID == playerID
	default:
		return invite.ToPlayerID == playerID || invite.FromPlayerID == playerID
	}
}
