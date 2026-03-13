package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/guild/internal/domain"
)

const (
	inviteDomainGuild = "guild"
	roleOwner         = "owner"
	roleMember        = "member"
)

// Invite contains the subset of invite state guild depends on.
type Invite struct {
	ID           string `json:"id"`
	Domain       string `json:"domain"`
	ResourceID   string `json:"resource_id,omitempty"`
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
	Status       string `json:"status"`
}

// InviteClient is the explicit boundary from guild to invite.
type InviteClient interface {
	CreateInvite(ctx context.Context, domainName string, resourceID string, fromPlayerID string, toPlayerID string) (Invite, *apperrors.Error)
	GetInvite(ctx context.Context, inviteID string) (Invite, *apperrors.Error)
}

// GuildService provides an in-memory prototype for guild creation and joins.
type GuildService struct {
	mu         sync.RWMutex
	guilds     map[string]domain.Guild
	invites    InviteClient
	now        func() time.Time
	newGuildID func() (string, error)
}

// NewGuildService constructs an in-memory guild service.
func NewGuildService(invites InviteClient) *GuildService {
	return &GuildService{
		guilds:  make(map[string]domain.Guild),
		invites: invites,
		now:     time.Now,
		newGuildID: func() (string, error) {
			return idgen.Token(8)
		},
	}
}

// CreateGuild creates a guild with an owner member.
func (s *GuildService) CreateGuild(name string, ownerID string) (domain.Guild, *apperrors.Error) {
	name = strings.TrimSpace(name)
	if name == "" || ownerID == "" {
		err := apperrors.New("invalid_request", "name and owner_id are required", http.StatusBadRequest)
		return domain.Guild{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	guildID, idErr := s.newGuildID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}

	now := s.now()
	guild := domain.Guild{
		ID:      guildID,
		Name:    name,
		OwnerID: ownerID,
		Members: []domain.GuildMember{
			{
				PlayerID: ownerID,
				Role:     roleOwner,
				JoinedAt: now,
			},
		},
		CreatedAt: now,
	}

	s.guilds[guild.ID] = guild
	return guild, nil
}

// GetGuild returns a guild by id.
func (s *GuildService) GetGuild(guildID string) (domain.Guild, *apperrors.Error) {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return domain.Guild{}, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	guild, ok := s.guilds[guildID]
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return domain.Guild{}, &err
	}

	return guild, nil
}

// CreateInvite issues a guild invite through the shared invite boundary.
func (s *GuildService) CreateInvite(ctx context.Context, guildID string, actorPlayerID string, toPlayerID string) (Invite, *apperrors.Error) {
	if guildID == "" || actorPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "guild_id, actor_player_id, and to_player_id are required", http.StatusBadRequest)
		return Invite{}, &err
	}

	s.mu.RLock()
	guild, ok := s.guilds[guildID]
	s.mu.RUnlock()
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return Invite{}, &err
	}

	if guild.OwnerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the guild owner can invite in the prototype", http.StatusForbidden)
		return Invite{}, &err
	}

	if hasMember(guild.Members, toPlayerID) {
		err := apperrors.New("already_member", "player is already in the guild", http.StatusConflict)
		return Invite{}, &err
	}

	return s.invites.CreateInvite(ctx, inviteDomainGuild, guildID, actorPlayerID, toPlayerID)
}

// JoinWithInvite adds a member after invite acceptance.
func (s *GuildService) JoinWithInvite(ctx context.Context, guildID string, inviteID string, actorPlayerID string) (domain.Guild, *apperrors.Error) {
	if guildID == "" || inviteID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "guild_id, invite_id, and actor_player_id are required", http.StatusBadRequest)
		return domain.Guild{}, &err
	}

	invite, appErr := s.invites.GetInvite(ctx, inviteID)
	if appErr != nil {
		return domain.Guild{}, appErr
	}

	if invite.Domain != inviteDomainGuild || invite.ResourceID != guildID {
		err := apperrors.New("forbidden", "invite does not belong to this guild", http.StatusForbidden)
		return domain.Guild{}, &err
	}

	if invite.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "invite belongs to a different player", http.StatusForbidden)
		return domain.Guild{}, &err
	}

	if invite.Status != "accepted" {
		err := apperrors.New("invite_not_accepted", "invite must be accepted before joining", http.StatusConflict)
		return domain.Guild{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	guild, ok := s.guilds[guildID]
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return domain.Guild{}, &err
	}

	if hasMember(guild.Members, actorPlayerID) {
		return guild, nil
	}

	guild.Members = append(guild.Members, domain.GuildMember{
		PlayerID: actorPlayerID,
		Role:     roleMember,
		JoinedAt: s.now(),
	})
	slices.SortFunc(guild.Members, func(a domain.GuildMember, b domain.GuildMember) int {
		switch {
		case a.PlayerID < b.PlayerID:
			return -1
		case a.PlayerID > b.PlayerID:
			return 1
		default:
			return 0
		}
	})

	s.guilds[guild.ID] = guild
	return guild, nil
}

func hasMember(members []domain.GuildMember, playerID string) bool {
	for _, member := range members {
		if member.PlayerID == playerID {
			return true
		}
	}

	return false
}

func (s *GuildService) String() string {
	return fmt.Sprintf("guild-service(guilds=%d)", len(s.guilds))
}
