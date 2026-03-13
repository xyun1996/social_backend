package service

import (
	"context"
	"fmt"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// PresenceRecord is the operator-facing player presence shape.
type PresenceRecord struct {
	PlayerID        string `json:"player_id"`
	Status          string `json:"status"`
	SessionID       string `json:"session_id"`
	RealmID         string `json:"realm_id,omitempty"`
	Location        string `json:"location,omitempty"`
	LastHeartbeatAt string `json:"last_heartbeat_at,omitempty"`
	LastSeenAt      string `json:"last_seen_at,omitempty"`
	ConnectedAt     string `json:"connected_at,omitempty"`
	DisconnectedAt  string `json:"disconnected_at,omitempty"`
}

// PartyMemberState is the operator-facing party member runtime shape.
type PartyMemberState struct {
	PlayerID  string `json:"player_id"`
	IsLeader  bool   `json:"is_leader"`
	IsReady   bool   `json:"is_ready"`
	Presence  string `json:"presence"`
	SessionID string `json:"session_id,omitempty"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// GuildMemberState is the operator-facing guild member runtime shape.
type GuildMemberState struct {
	PlayerID  string `json:"player_id"`
	Role      string `json:"role"`
	Presence  string `json:"presence"`
	SessionID string `json:"session_id,omitempty"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// PartySnapshot aggregates current party runtime state.
type PartySnapshot struct {
	PartyID string             `json:"party_id"`
	Count   int                `json:"count"`
	Members []PartyMemberState `json:"members"`
}

// GuildSnapshot aggregates current guild runtime state.
type GuildSnapshot struct {
	GuildID string             `json:"guild_id"`
	Count   int                `json:"count"`
	Members []GuildMemberState `json:"members"`
}

// PresenceReader exposes the presence read boundary for ops.
type PresenceReader interface {
	GetPresence(ctx context.Context, playerID string) (PresenceRecord, *apperrors.Error)
}

// PartyReader exposes the party read boundary for ops.
type PartyReader interface {
	GetPartySnapshot(ctx context.Context, partyID string) (PartySnapshot, *apperrors.Error)
}

// GuildReader exposes the guild read boundary for ops.
type GuildReader interface {
	GetGuildSnapshot(ctx context.Context, guildID string) (GuildSnapshot, *apperrors.Error)
}

// OpsService provides operator-facing read aggregation.
type OpsService struct {
	presence PresenceReader
	parties  PartyReader
	guilds   GuildReader
}

// NewOpsService constructs the operator read service.
func NewOpsService(presence PresenceReader, parties PartyReader, guilds GuildReader) *OpsService {
	return &OpsService{
		presence: presence,
		parties:  parties,
		guilds:   guilds,
	}
}

// GetPlayerPresence returns the operator-facing presence view.
func (s *OpsService) GetPlayerPresence(ctx context.Context, playerID string) (PresenceRecord, *apperrors.Error) {
	if s.presence == nil {
		err := apperrors.New("dependency_missing", "presence reader is not configured", 500)
		return PresenceRecord{}, &err
	}
	return s.presence.GetPresence(ctx, playerID)
}

// GetPartySnapshot returns the operator-facing party snapshot.
func (s *OpsService) GetPartySnapshot(ctx context.Context, partyID string) (PartySnapshot, *apperrors.Error) {
	if s.parties == nil {
		err := apperrors.New("dependency_missing", "party reader is not configured", 500)
		return PartySnapshot{}, &err
	}
	return s.parties.GetPartySnapshot(ctx, partyID)
}

// GetGuildSnapshot returns the operator-facing guild snapshot.
func (s *OpsService) GetGuildSnapshot(ctx context.Context, guildID string) (GuildSnapshot, *apperrors.Error) {
	if s.guilds == nil {
		err := apperrors.New("dependency_missing", "guild reader is not configured", 500)
		return GuildSnapshot{}, &err
	}
	return s.guilds.GetGuildSnapshot(ctx, guildID)
}

func (s *OpsService) String() string {
	return fmt.Sprintf("ops-service(presence=%t,party=%t,guild=%t)", s.presence != nil, s.parties != nil, s.guilds != nil)
}
