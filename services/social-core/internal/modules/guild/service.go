package guild

import (
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	invitemodule "github.com/xyun1996/social_backend/services/social-core/internal/modules/invite"
)

const (
	roleOwner  = "owner"
	roleMember = "member"
)

type Member struct {
	PlayerID string    `json:"player_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Guild struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	OwnerID               string    `json:"owner_id"`
	Announcement          string    `json:"announcement,omitempty"`
	AnnouncementUpdatedAt time.Time `json:"announcement_updated_at,omitempty"`
	Members               []Member  `json:"members"`
	CreatedAt             time.Time `json:"created_at"`
}

type LogEntry struct {
	ID        string    `json:"id"`
	GuildID   string    `json:"guild_id"`
	Action    string    `json:"action"`
	ActorID   string    `json:"actor_id,omitempty"`
	TargetID  string    `json:"target_id,omitempty"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type InviteBoundary interface {
	CreateInvite(domainName, resourceID, fromPlayerID, toPlayerID string, ttl time.Duration) (invitemodule.Invite, *apperrors.Error)
	GetInvite(inviteID string) (invitemodule.Invite, *apperrors.Error)
}

type Service struct {
	mu      sync.RWMutex
	now     func() time.Time
	invites InviteBoundary
	guilds  map[string]Guild
	logs    map[string][]LogEntry
}

func NewService(invites InviteBoundary) *Service {
	return &Service{
		now:     time.Now,
		invites: invites,
		guilds:  make(map[string]Guild),
		logs:    make(map[string][]LogEntry),
	}
}

func (s *Service) CreateGuild(name, ownerID string) (Guild, *apperrors.Error) {
	name = strings.TrimSpace(name)
	if name == "" || ownerID == "" {
		err := apperrors.New("invalid_request", "name and owner_id are required", http.StatusBadRequest)
		return Guild{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, guild := range s.guilds {
		if hasMember(guild.Members, ownerID) {
			err := apperrors.New("already_in_guild", "owner already belongs to a guild", http.StatusConflict)
			return Guild{}, &err
		}
	}

	guildID, err := idgen.Token(8)
	if err != nil {
		internal := apperrors.Internal()
		return Guild{}, &internal
	}
	now := s.now()
	guild := Guild{
		ID:        guildID,
		Name:      name,
		OwnerID:   ownerID,
		Members:   []Member{{PlayerID: ownerID, Role: roleOwner, JoinedAt: now}},
		CreatedAt: now,
	}
	s.guilds[guild.ID] = guild
	s.appendLogLocked(guild.ID, "guild.created", ownerID, "", "guild created")
	return guild, nil
}

func (s *Service) GetGuild(guildID string) (Guild, *apperrors.Error) {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return Guild{}, &err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	guild, ok := s.guilds[guildID]
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return Guild{}, &err
	}
	return guild, nil
}

func (s *Service) FindGuildByPlayer(playerID string) (Guild, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return Guild{}, &err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, guild := range s.guilds {
		if hasMember(guild.Members, playerID) {
			return guild, nil
		}
	}
	err := apperrors.New("not_found", "guild not found for player", http.StatusNotFound)
	return Guild{}, &err
}

func (s *Service) ListMembers(guildID string) ([]Member, *apperrors.Error) {
	guild, appErr := s.GetGuild(guildID)
	if appErr != nil {
		return nil, appErr
	}
	return guild.Members, nil
}

func (s *Service) ListLogs(guildID string) ([]LogEntry, *apperrors.Error) {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return nil, &err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.guilds[guildID]; !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return nil, &err
	}
	return append([]LogEntry(nil), s.logs[guildID]...), nil
}

func (s *Service) UpdateAnnouncement(guildID, actorPlayerID, announcement string) (Guild, *apperrors.Error) {
	if guildID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "guild_id and actor_player_id are required", http.StatusBadRequest)
		return Guild{}, &err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	guild, ok := s.guilds[guildID]
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return Guild{}, &err
	}
	if guild.OwnerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the guild owner can update the announcement", http.StatusForbidden)
		return Guild{}, &err
	}
	guild.Announcement = strings.TrimSpace(announcement)
	guild.AnnouncementUpdatedAt = s.now()
	s.guilds[guildID] = guild
	s.appendLogLocked(guild.ID, "guild.announcement_updated", actorPlayerID, "", "guild announcement updated")
	return guild, nil
}

func (s *Service) CreateInvite(guildID, actorPlayerID, toPlayerID string) (invitemodule.Invite, *apperrors.Error) {
	if guildID == "" || actorPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "guild_id, actor_player_id, and to_player_id are required", http.StatusBadRequest)
		return invitemodule.Invite{}, &err
	}
	s.mu.RLock()
	guild, ok := s.guilds[guildID]
	s.mu.RUnlock()
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return invitemodule.Invite{}, &err
	}
	if guild.OwnerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the guild owner can invite in phase A", http.StatusForbidden)
		return invitemodule.Invite{}, &err
	}
	if hasMember(guild.Members, toPlayerID) {
		err := apperrors.New("already_member", "player is already in the guild", http.StatusConflict)
		return invitemodule.Invite{}, &err
	}
	return s.invites.CreateInvite("guild", guildID, actorPlayerID, toPlayerID, 0)
}

func (s *Service) JoinWithInvite(guildID, inviteID, actorPlayerID string) (Guild, *apperrors.Error) {
	if guildID == "" || inviteID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "guild_id, invite_id, and actor_player_id are required", http.StatusBadRequest)
		return Guild{}, &err
	}
	invite, appErr := s.invites.GetInvite(inviteID)
	if appErr != nil {
		return Guild{}, appErr
	}
	if invite.Domain != "guild" || invite.ResourceID != guildID {
		err := apperrors.New("forbidden", "invite does not belong to this guild", http.StatusForbidden)
		return Guild{}, &err
	}
	if invite.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "invite belongs to a different player", http.StatusForbidden)
		return Guild{}, &err
	}
	if invite.Status != invitemodule.StatusAccepted {
		err := apperrors.New("invite_not_accepted", "invite must be accepted before joining", http.StatusConflict)
		return Guild{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	guild, ok := s.guilds[guildID]
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return Guild{}, &err
	}
	if hasMember(guild.Members, actorPlayerID) {
		return guild, nil
	}
	guild.Members = append(guild.Members, Member{PlayerID: actorPlayerID, Role: roleMember, JoinedAt: s.now()})
	slices.SortFunc(guild.Members, func(a, b Member) int {
		if a.PlayerID < b.PlayerID {
			return -1
		}
		if a.PlayerID > b.PlayerID {
			return 1
		}
		return 0
	})
	s.guilds[guildID] = guild
	s.appendLogLocked(guild.ID, "guild.joined", actorPlayerID, actorPlayerID, "member joined guild")
	return guild, nil
}

func hasMember(members []Member, playerID string) bool {
	for _, member := range members {
		if member.PlayerID == playerID {
			return true
		}
	}
	return false
}

func (s *Service) appendLogLocked(guildID, action, actorID, targetID, message string) {
	logID, err := idgen.Token(10)
	if err != nil {
		return
	}
	entry := LogEntry{
		ID:        logID,
		GuildID:   guildID,
		Action:    action,
		ActorID:   actorID,
		TargetID:  targetID,
		Message:   message,
		CreatedAt: s.now(),
	}
	s.logs[guildID] = append(s.logs[guildID], entry)
}
