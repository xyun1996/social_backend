package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/guild/internal/domain"
)

const (
	inviteDomainGuild = "guild"
	roleOwner         = "owner"
	roleMember        = "member"
	presenceOnline    = "online"
	presenceOffline   = "offline"
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

// PresenceSnapshot contains the subset of presence state guild uses.
type PresenceSnapshot struct {
	PlayerID  string `json:"player_id"`
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// PresenceReader resolves presence state for guild member views.
type PresenceReader interface {
	GetPresence(ctx context.Context, playerID string) (PresenceSnapshot, *apperrors.Error)
}

// MemberState combines guild role and presence state.
type MemberState struct {
	PlayerID  string `json:"player_id"`
	Role      string `json:"role"`
	Presence  string `json:"presence"`
	SessionID string `json:"session_id,omitempty"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// GuildService provides an in-memory prototype for guild creation and joins.
type GuildService struct {
	guilds     GuildStore
	invites    InviteClient
	presence   PresenceReader
	now        func() time.Time
	newGuildID func() (string, error)
}

// NewGuildService constructs an in-memory guild service.
func NewGuildService(invites InviteClient, presence PresenceReader) *GuildService {
	return NewGuildServiceWithStore(newMemoryGuildStore(), invites, presence)
}

// NewGuildServiceWithStore constructs a guild service with injected persistence boundaries.
func NewGuildServiceWithStore(guilds GuildStore, invites InviteClient, presence PresenceReader) *GuildService {
	return &GuildService{
		guilds:   guilds,
		invites:  invites,
		presence: presence,
		now:      time.Now,
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

	if err := s.guilds.SaveGuild(guild); err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
	return guild, nil
}

// ListMemberStates returns role and presence state for current guild members.
func (s *GuildService) ListMemberStates(ctx context.Context, guildID string) ([]MemberState, *apperrors.Error) {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return nil, &err
	}

	guild, ok, err := s.guilds.GetGuild(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return nil, &err
	}

	states := make([]MemberState, 0, len(guild.Members))
	for _, member := range guild.Members {
		state := MemberState{
			PlayerID: member.PlayerID,
			Role:     member.Role,
			Presence: presenceOffline,
		}

		if s.presence != nil {
			snapshot, appErr := s.presence.GetPresence(ctx, member.PlayerID)
			if appErr != nil && appErr.Code != "not_found" {
				return nil, appErr
			}
			if appErr == nil {
				state.Presence = snapshot.Status
				state.SessionID = snapshot.SessionID
				state.RealmID = snapshot.RealmID
				state.Location = snapshot.Location
			}
		}

		states = append(states, state)
	}

	return states, nil
}

// GetGuild returns a guild by id.
func (s *GuildService) GetGuild(guildID string) (domain.Guild, *apperrors.Error) {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return domain.Guild{}, &err
	}

	guild, ok, err := s.guilds.GetGuild(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
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

	guild, ok, err := s.guilds.GetGuild(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return Invite{}, &internal
	}
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

	guild, ok, err := s.guilds.GetGuild(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
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

	if err := s.guilds.SaveGuild(guild); err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
	return guild, nil
}

// KickMember removes a non-owner member at the direction of the guild owner.
func (s *GuildService) KickMember(guildID string, actorPlayerID string, targetPlayerID string) (domain.Guild, *apperrors.Error) {
	if guildID == "" || actorPlayerID == "" || targetPlayerID == "" {
		err := apperrors.New("invalid_request", "guild_id, actor_player_id, and target_player_id are required", http.StatusBadRequest)
		return domain.Guild{}, &err
	}

	guild, ok, err := s.guilds.GetGuild(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return domain.Guild{}, &err
	}
	if guild.OwnerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the guild owner can kick members", http.StatusForbidden)
		return domain.Guild{}, &err
	}
	if targetPlayerID == guild.OwnerID {
		err := apperrors.New("invalid_request", "guild owner cannot kick themselves", http.StatusBadRequest)
		return domain.Guild{}, &err
	}
	if !hasMember(guild.Members, targetPlayerID) {
		err := apperrors.New("not_found", "target member not found", http.StatusNotFound)
		return domain.Guild{}, &err
	}

	guild.Members = deleteGuildMember(guild.Members, targetPlayerID)
	if err := s.guilds.SaveGuild(guild); err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
	return guild, nil
}

// TransferOwnership transfers guild ownership to another current member.
func (s *GuildService) TransferOwnership(guildID string, actorPlayerID string, targetPlayerID string) (domain.Guild, *apperrors.Error) {
	if guildID == "" || actorPlayerID == "" || targetPlayerID == "" {
		err := apperrors.New("invalid_request", "guild_id, actor_player_id, and target_player_id are required", http.StatusBadRequest)
		return domain.Guild{}, &err
	}

	guild, ok, err := s.guilds.GetGuild(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return domain.Guild{}, &err
	}
	if guild.OwnerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the guild owner can transfer ownership", http.StatusForbidden)
		return domain.Guild{}, &err
	}
	if !hasMember(guild.Members, targetPlayerID) {
		err := apperrors.New("not_found", "target member not found", http.StatusNotFound)
		return domain.Guild{}, &err
	}

	for i := range guild.Members {
		switch guild.Members[i].PlayerID {
		case actorPlayerID:
			guild.Members[i].Role = roleMember
		case targetPlayerID:
			guild.Members[i].Role = roleOwner
		}
	}
	guild.OwnerID = targetPlayerID
	if err := s.guilds.SaveGuild(guild); err != nil {
		internal := apperrors.Internal()
		return domain.Guild{}, &internal
	}
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

func deleteGuildMember(members []domain.GuildMember, playerID string) []domain.GuildMember {
	filtered := members[:0]
	for _, member := range members {
		if member.PlayerID == playerID {
			continue
		}
		filtered = append(filtered, member)
	}
	return filtered
}

func (s *GuildService) String() string {
	guilds, err := s.guilds.ListGuilds()
	if err != nil {
		return "guild-service(guilds=unknown)"
	}
	return fmt.Sprintf("guild-service(guilds=%d)", len(guilds))
}
