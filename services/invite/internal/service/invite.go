package service

import (
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/invite/internal/domain"
)

const (
	inviteStatusPending  = "pending"
	inviteStatusAccepted = "accepted"
	inviteStatusDeclined = "declined"
	inviteStatusExpired  = "expired"

	inviteActionAccept  = "accept"
	inviteActionDecline = "decline"

	defaultInviteTTL = 15 * time.Minute
)

// InviteService provides an in-memory prototype for cross-domain invite flows.
type InviteService struct {
	mu          sync.RWMutex
	invites     map[string]domain.Invite
	now         func() time.Time
	newInviteID func() (string, error)
}

// NewInviteService constructs an in-memory invite service.
func NewInviteService() *InviteService {
	return &InviteService{
		invites: make(map[string]domain.Invite),
		now:     time.Now,
		newInviteID: func() (string, error) {
			return idgen.Token(8)
		},
	}
}

// CreateInvite creates or returns an existing pending invite for the same tuple.
func (s *InviteService) CreateInvite(domainName string, resourceID string, fromPlayerID string, toPlayerID string, ttl time.Duration) (domain.Invite, *apperrors.Error) {
	if domainName == "" || fromPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "domain, from_player_id, and to_player_id are required", http.StatusBadRequest)
		return domain.Invite{}, &err
	}

	if fromPlayerID == toPlayerID {
		err := apperrors.New("invalid_request", "cannot invite yourself", http.StatusBadRequest)
		return domain.Invite{}, &err
	}

	if ttl <= 0 {
		ttl = defaultInviteTTL
	}

	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.expirePendingLocked(now)

	for _, invite := range s.invites {
		if invite.Domain == domainName &&
			invite.ResourceID == resourceID &&
			invite.FromPlayerID == fromPlayerID &&
			invite.ToPlayerID == toPlayerID &&
			invite.Status == inviteStatusPending {
			return invite, nil
		}
	}

	inviteID, err := s.newInviteID()
	if err != nil {
		internal := apperrors.Internal()
		return domain.Invite{}, &internal
	}

	invite := domain.Invite{
		ID:           inviteID,
		Domain:       domainName,
		ResourceID:   resourceID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		Status:       inviteStatusPending,
		CreatedAt:    now,
		ExpiresAt:    now.Add(ttl),
	}

	s.invites[invite.ID] = invite
	return invite, nil
}

// GetInvite returns a stored invite by id.
func (s *InviteService) GetInvite(inviteID string) (domain.Invite, *apperrors.Error) {
	if inviteID == "" {
		err := apperrors.New("invalid_request", "invite_id is required", http.StatusBadRequest)
		return domain.Invite{}, &err
	}

	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()

	invite, ok := s.invites[inviteID]
	if !ok {
		err := apperrors.New("not_found", "invite not found", http.StatusNotFound)
		return domain.Invite{}, &err
	}

	if invite.Status == inviteStatusPending && s.isExpired(invite, now) {
		invite.Status = inviteStatusExpired
		s.invites[invite.ID] = invite
	}

	return invite, nil
}

// RespondInvite accepts or declines a pending invite.
func (s *InviteService) RespondInvite(inviteID string, actorPlayerID string, action string) (domain.Invite, *apperrors.Error) {
	if inviteID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "invite_id and actor_player_id are required", http.StatusBadRequest)
		return domain.Invite{}, &err
	}

	if action != inviteActionAccept && action != inviteActionDecline {
		err := apperrors.New("invalid_request", "action must be accept or decline", http.StatusBadRequest)
		return domain.Invite{}, &err
	}

	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()

	invite, ok := s.invites[inviteID]
	if !ok {
		err := apperrors.New("not_found", "invite not found", http.StatusNotFound)
		return domain.Invite{}, &err
	}

	if invite.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the invited player can respond", http.StatusForbidden)
		return domain.Invite{}, &err
	}

	if s.isExpired(invite, now) {
		invite.Status = inviteStatusExpired
		s.invites[invite.ID] = invite
		err := apperrors.New("invite_expired", "invite has expired", http.StatusConflict)
		return invite, &err
	}

	switch invite.Status {
	case inviteStatusAccepted, inviteStatusDeclined:
		return invite, nil
	case inviteStatusExpired:
		err := apperrors.New("invite_expired", "invite has expired", http.StatusConflict)
		return invite, &err
	}

	invite.RespondedAt = &now
	if action == inviteActionAccept {
		invite.Status = inviteStatusAccepted
	} else {
		invite.Status = inviteStatusDeclined
	}

	s.invites[invite.ID] = invite
	return invite, nil
}

// ListInvites returns invites involving a player, optionally filtered by role and status.
func (s *InviteService) ListInvites(playerID string, role string, status string) ([]domain.Invite, *apperrors.Error) {
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

	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.expirePendingLocked(now)

	invites := make([]domain.Invite, 0)
	for _, invite := range s.invites {
		if !matchesRole(invite, playerID, role) {
			continue
		}

		if status != "" && invite.Status != status {
			continue
		}

		invites = append(invites, invite)
	}

	slices.SortFunc(invites, func(a domain.Invite, b domain.Invite) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}

		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})

	return invites, nil
}

func matchesRole(invite domain.Invite, playerID string, role string) bool {
	switch role {
	case "inbox":
		return invite.ToPlayerID == playerID
	case "outbox":
		return invite.FromPlayerID == playerID
	default:
		return invite.FromPlayerID == playerID || invite.ToPlayerID == playerID
	}
}

func (s *InviteService) expirePendingLocked(now time.Time) {
	for id, invite := range s.invites {
		if invite.Status != inviteStatusPending {
			continue
		}

		if s.isExpired(invite, now) {
			invite.Status = inviteStatusExpired
			s.invites[id] = invite
		}
	}
}

func (s *InviteService) isExpired(invite domain.Invite, now time.Time) bool {
	return !invite.ExpiresAt.After(now)
}

func (s *InviteService) String() string {
	return fmt.Sprintf("invite-service(invites=%d)", len(s.invites))
}
