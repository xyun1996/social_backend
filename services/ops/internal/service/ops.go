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

// SocialSnapshot aggregates current social relationship state for a player.
type SocialSnapshot struct {
	PlayerID string   `json:"player_id"`
	Friends  []string `json:"friends"`
	Blocks   []string `json:"blocks"`
}

// PlayerOverview aggregates the operator-facing player runtime state.
type PlayerOverview struct {
	PlayerID  string         `json:"player_id"`
	Presence  PresenceRecord `json:"presence"`
	Friends   []string       `json:"friends"`
	Blocks    []string       `json:"blocks"`
	FriendCnt int            `json:"friend_count"`
	BlockCnt  int            `json:"block_count"`
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

// WorkerJob is the operator-facing async job shape.
type WorkerJob struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Payload     string `json:"payload"`
	Status      string `json:"status"`
	Attempts    int    `json:"attempts"`
	LastError   string `json:"last_error,omitempty"`
	ClaimedBy   string `json:"claimed_by,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	ClaimedAt   string `json:"claimed_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// WorkerSnapshot aggregates async job queue state.
type WorkerSnapshot struct {
	Status string      `json:"status,omitempty"`
	Type   string      `json:"type,omitempty"`
	Count  int         `json:"count"`
	Jobs   []WorkerJob `json:"jobs"`
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

// WorkerReader exposes the worker read boundary for ops.
type WorkerReader interface {
	GetWorkerSnapshot(ctx context.Context, status string, jobType string) (WorkerSnapshot, *apperrors.Error)
}

// SocialReader exposes the social read boundary for ops.
type SocialReader interface {
	GetSocialSnapshot(ctx context.Context, playerID string) (SocialSnapshot, *apperrors.Error)
}

// OpsService provides operator-facing read aggregation.
type OpsService struct {
	presence PresenceReader
	parties  PartyReader
	guilds   GuildReader
	worker   WorkerReader
	social   SocialReader
}

// NewOpsService constructs the operator read service.
func NewOpsService(presence PresenceReader, parties PartyReader, guilds GuildReader, worker WorkerReader, social SocialReader) *OpsService {
	return &OpsService{
		presence: presence,
		parties:  parties,
		guilds:   guilds,
		worker:   worker,
		social:   social,
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

// GetWorkerSnapshot returns the operator-facing worker snapshot.
func (s *OpsService) GetWorkerSnapshot(ctx context.Context, status string, jobType string) (WorkerSnapshot, *apperrors.Error) {
	if s.worker == nil {
		err := apperrors.New("dependency_missing", "worker reader is not configured", 500)
		return WorkerSnapshot{}, &err
	}
	return s.worker.GetWorkerSnapshot(ctx, status, jobType)
}

// GetPlayerOverview returns the operator-facing player runtime overview.
func (s *OpsService) GetPlayerOverview(ctx context.Context, playerID string) (PlayerOverview, *apperrors.Error) {
	if s.presence == nil {
		err := apperrors.New("dependency_missing", "presence reader is not configured", 500)
		return PlayerOverview{}, &err
	}
	if s.social == nil {
		err := apperrors.New("dependency_missing", "social reader is not configured", 500)
		return PlayerOverview{}, &err
	}

	presence, appErr := s.presence.GetPresence(ctx, playerID)
	if appErr != nil {
		return PlayerOverview{}, appErr
	}

	social, appErr := s.social.GetSocialSnapshot(ctx, playerID)
	if appErr != nil {
		return PlayerOverview{}, appErr
	}

	return PlayerOverview{
		PlayerID:  playerID,
		Presence:  presence,
		Friends:   social.Friends,
		Blocks:    social.Blocks,
		FriendCnt: len(social.Friends),
		BlockCnt:  len(social.Blocks),
	}, nil
}

func (s *OpsService) String() string {
	return fmt.Sprintf("ops-service(presence=%t,party=%t,guild=%t,worker=%t,social=%t)", s.presence != nil, s.parties != nil, s.guilds != nil, s.worker != nil, s.social != nil)
}
