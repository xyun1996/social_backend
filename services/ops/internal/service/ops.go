package service

import (
	"context"
	"fmt"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

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

type SocialRelationshipDetail struct {
	TargetPlayerID   string `json:"target_player_id"`
	State            string `json:"state"`
	IsFriend         bool   `json:"is_friend"`
	HasPendingInbox  bool   `json:"has_pending_inbox"`
	HasPendingOutbox bool   `json:"has_pending_outbox"`
	IsBlocked        bool   `json:"is_blocked"`
	IsBlockedBy      bool   `json:"is_blocked_by"`
	Remark           string `json:"remark,omitempty"`
	ReverseRemark    string `json:"reverse_remark,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type SocialSnapshot struct {
	PlayerID            string                     `json:"player_id"`
	Friends             []string                   `json:"friends"`
	Blocks              []string                   `json:"blocks"`
	PendingInbox        []string                   `json:"pending_inbox"`
	PendingOutbox       []string                   `json:"pending_outbox"`
	PendingTotal        int                        `json:"pending_total"`
	RelationshipDetails []SocialRelationshipDetail `json:"relationship_details,omitempty"`
}

type PlayerOverview struct {
	PlayerID              string                     `json:"player_id"`
	Presence              PresenceRecord             `json:"presence"`
	Friends               []string                   `json:"friends"`
	Blocks                []string                   `json:"blocks"`
	PendingInbox          []string                   `json:"pending_inbox"`
	PendingOutbox         []string                   `json:"pending_outbox"`
	FriendCnt             int                        `json:"friend_count"`
	BlockCnt              int                        `json:"block_count"`
	PendingInboxCount     int                        `json:"pending_inbox_count"`
	PendingOutboxCount    int                        `json:"pending_outbox_count"`
	RelationshipCount     int                        `json:"relationship_count"`
	RelationshipDetails   []SocialRelationshipDetail `json:"relationship_details,omitempty"`
	CurrentPartyID        string                     `json:"current_party_id,omitempty"`
	CurrentGuildID        string                     `json:"current_guild_id,omitempty"`
	CurrentGuildRole      string                     `json:"current_guild_role,omitempty"`
	CurrentQueueStatus    string                     `json:"current_queue_status,omitempty"`
	CurrentQueueExpiresAt string                     `json:"current_queue_expires_at,omitempty"`
}

type PartyMemberState struct { PlayerID string `json:"player_id"`; IsLeader bool `json:"is_leader"`; IsReady bool `json:"is_ready"`; Presence string `json:"presence"`; SessionID string `json:"session_id,omitempty"`; RealmID string `json:"realm_id,omitempty"`; Location string `json:"location,omitempty"` }
type PartyQueueState struct { PartyID string `json:"party_id"`; QueueName string `json:"queue_name"`; Status string `json:"status"`; JoinedBy string `json:"joined_by"`; JoinedAt string `json:"joined_at,omitempty"`; ExpiresAt string `json:"expires_at,omitempty"` }
type GuildMemberState struct { PlayerID string `json:"player_id"`; Role string `json:"role"`; Presence string `json:"presence"`; SessionID string `json:"session_id,omitempty"`; RealmID string `json:"realm_id,omitempty"`; Location string `json:"location,omitempty"` }
type GuildLogEntry struct { ID string `json:"id"`; Action string `json:"action"`; ActorID string `json:"actor_id,omitempty"`; TargetID string `json:"target_id,omitempty"`; Message string `json:"message,omitempty"`; CreatedAt string `json:"created_at,omitempty"` }
type GuildContribution struct { PlayerID string `json:"player_id"`; TotalXP int `json:"total_xp"`; LastSourceType string `json:"last_source_type,omitempty"`; UpdatedAt string `json:"updated_at,omitempty"` }
type GuildActivityInstance struct { ID string `json:"id"`; TemplateKey string `json:"template_key"`; PeriodKey string `json:"period_key"`; Status string `json:"status"`; StartsAt string `json:"starts_at,omitempty"`; EndsAt string `json:"ends_at,omitempty"`; UpdatedAt string `json:"updated_at,omitempty"` }
type GuildRewardRecord struct { ID string `json:"id"`; PlayerID string `json:"player_id"`; ActivityID string `json:"activity_id"`; TemplateKey string `json:"template_key"`; RewardType string `json:"reward_type"`; RewardRef string `json:"reward_ref,omitempty"`; CreatedAt string `json:"created_at,omitempty"` }
type PartySnapshot struct { PartyID string `json:"party_id"`; Count int `json:"count"`; Members []PartyMemberState `json:"members"`; Queue *PartyQueueState `json:"queue,omitempty"` }
type GuildSnapshot struct { GuildID string `json:"guild_id"`; Name string `json:"name,omitempty"`; OwnerID string `json:"owner_id,omitempty"`; Announcement string `json:"announcement,omitempty"`; AnnouncementUpdatedAt string `json:"announcement_updated_at,omitempty"`; Level int `json:"level"`; Experience int `json:"experience"`; NextLevelXP int `json:"next_level_xp"`; Count int `json:"count"`; Members []GuildMemberState `json:"members"`; LogCount int `json:"log_count"`; Logs []GuildLogEntry `json:"logs,omitempty"`; Contributions []GuildContribution `json:"contributions,omitempty"`; ActivityInstances []GuildActivityInstance `json:"activity_instances,omitempty"`; RewardRecords []GuildRewardRecord `json:"reward_records,omitempty"` }
type WorkerJob struct { ID string `json:"id"`; Type string `json:"type"`; Payload string `json:"payload"`; Status string `json:"status"`; Attempts int `json:"attempts"`; MaxAttempts int `json:"max_attempts"`; LastError string `json:"last_error,omitempty"`; ClaimedBy string `json:"claimed_by,omitempty"`; CreatedAt string `json:"created_at,omitempty"`; ClaimedAt string `json:"claimed_at,omitempty"`; CompletedAt string `json:"completed_at,omitempty"`; NextAttemptAt string `json:"next_attempt_at,omitempty"` }
type WorkerSnapshot struct { Status string `json:"status,omitempty"`; Type string `json:"type,omitempty"`; Count int `json:"count"`; Jobs []WorkerJob `json:"jobs"` }
type MySQLBootstrapService struct { Service string `json:"service"`; Count int `json:"count"`; MigrationIDs []string `json:"migration_ids"` }
type MySQLBootstrapSnapshot struct { Count int `json:"count"`; Services []MySQLBootstrapService `json:"services"` }
type RedisWorkerStatusCount struct { Status string `json:"status"`; Count int `json:"count"` }
type RedisRuntimeSnapshot struct { RedisURL string `json:"redis_url,omitempty"`; PresenceRecordCount int `json:"presence_record_count"`; GatewaySessionCount int `json:"gateway_session_count"`; WorkerJobCount int `json:"worker_job_count"`; WorkerStatusCounters []RedisWorkerStatusCount `json:"worker_status_counters"` }
type DurableSummary struct { MySQL *MySQLBootstrapSnapshot `json:"mysql,omitempty"`; Redis *RedisRuntimeSnapshot `json:"redis,omitempty"` }

type PresenceReader interface { GetPresence(ctx context.Context, playerID string) (PresenceRecord, *apperrors.Error) }
type PartyReader interface { GetPartySnapshot(ctx context.Context, partyID string) (PartySnapshot, *apperrors.Error); GetPartyByPlayer(ctx context.Context, playerID string) (PartySnapshot, *apperrors.Error) }
type GuildReader interface { GetGuildSnapshot(ctx context.Context, guildID string) (GuildSnapshot, *apperrors.Error); GetGuildByPlayer(ctx context.Context, playerID string) (GuildSnapshot, *apperrors.Error) }
type WorkerReader interface { GetWorkerSnapshot(ctx context.Context, status string, jobType string) (WorkerSnapshot, *apperrors.Error) }
type SocialReader interface { GetSocialSnapshot(ctx context.Context, playerID string) (SocialSnapshot, *apperrors.Error) }
type BootstrapReader interface { GetMySQLBootstrapSnapshot(ctx context.Context) (MySQLBootstrapSnapshot, *apperrors.Error) }
type RedisRuntimeReader interface { GetRedisRuntimeSnapshot(ctx context.Context) (RedisRuntimeSnapshot, *apperrors.Error) }

type OpsService struct { presence PresenceReader; parties PartyReader; guilds GuildReader; worker WorkerReader; social SocialReader; bootstrap BootstrapReader; redis RedisRuntimeReader }
func NewOpsService(presence PresenceReader, parties PartyReader, guilds GuildReader, worker WorkerReader, social SocialReader, bootstrap BootstrapReader, redis RedisRuntimeReader) *OpsService { return &OpsService{presence: presence, parties: parties, guilds: guilds, worker: worker, social: social, bootstrap: bootstrap, redis: redis} }
func (s *OpsService) GetPlayerPresence(ctx context.Context, playerID string) (PresenceRecord, *apperrors.Error) { if s.presence == nil { err := apperrors.New("dependency_missing", "presence reader is not configured", 500); return PresenceRecord{}, &err }; return s.presence.GetPresence(ctx, playerID) }
func (s *OpsService) GetPartySnapshot(ctx context.Context, partyID string) (PartySnapshot, *apperrors.Error) { if s.parties == nil { err := apperrors.New("dependency_missing", "party reader is not configured", 500); return PartySnapshot{}, &err }; return s.parties.GetPartySnapshot(ctx, partyID) }
func (s *OpsService) GetGuildSnapshot(ctx context.Context, guildID string) (GuildSnapshot, *apperrors.Error) { if s.guilds == nil { err := apperrors.New("dependency_missing", "guild reader is not configured", 500); return GuildSnapshot{}, &err }; return s.guilds.GetGuildSnapshot(ctx, guildID) }
func (s *OpsService) GetWorkerSnapshot(ctx context.Context, status string, jobType string) (WorkerSnapshot, *apperrors.Error) { if s.worker == nil { err := apperrors.New("dependency_missing", "worker reader is not configured", 500); return WorkerSnapshot{}, &err }; return s.worker.GetWorkerSnapshot(ctx, status, jobType) }
func (s *OpsService) GetSocialSnapshot(ctx context.Context, playerID string) (SocialSnapshot, *apperrors.Error) { if s.social == nil { err := apperrors.New("dependency_missing", "social reader is not configured", 500); return SocialSnapshot{}, &err }; return s.social.GetSocialSnapshot(ctx, playerID) }
func (s *OpsService) GetMySQLBootstrapSnapshot(ctx context.Context) (MySQLBootstrapSnapshot, *apperrors.Error) { if s.bootstrap == nil { err := apperrors.New("dependency_missing", "mysql bootstrap reader is not configured", 500); return MySQLBootstrapSnapshot{}, &err }; return s.bootstrap.GetMySQLBootstrapSnapshot(ctx) }
func (s *OpsService) GetRedisRuntimeSnapshot(ctx context.Context) (RedisRuntimeSnapshot, *apperrors.Error) { if s.redis == nil { err := apperrors.New("dependency_missing", "redis runtime reader is not configured", 500); return RedisRuntimeSnapshot{}, &err }; return s.redis.GetRedisRuntimeSnapshot(ctx) }
func (s *OpsService) GetDurableSummary(ctx context.Context) (DurableSummary, *apperrors.Error) { summary := DurableSummary{}; if s.bootstrap != nil { record, appErr := s.bootstrap.GetMySQLBootstrapSnapshot(ctx); if appErr != nil { return DurableSummary{}, appErr }; summary.MySQL = &record }; if s.redis != nil { record, appErr := s.redis.GetRedisRuntimeSnapshot(ctx); if appErr != nil { return DurableSummary{}, appErr }; summary.Redis = &record }; if summary.MySQL == nil && summary.Redis == nil { err := apperrors.New("dependency_missing", "no durable status readers are configured", 500); return DurableSummary{}, &err }; return summary, nil }

func (s *OpsService) GetPlayerOverview(ctx context.Context, playerID string) (PlayerOverview, *apperrors.Error) {
	if s.presence == nil { err := apperrors.New("dependency_missing", "presence reader is not configured", 500); return PlayerOverview{}, &err }
	if s.social == nil { err := apperrors.New("dependency_missing", "social reader is not configured", 500); return PlayerOverview{}, &err }
	presence, appErr := s.presence.GetPresence(ctx, playerID); if appErr != nil { return PlayerOverview{}, appErr }
	social, appErr := s.social.GetSocialSnapshot(ctx, playerID); if appErr != nil { return PlayerOverview{}, appErr }
	overview := PlayerOverview{PlayerID: playerID, Presence: presence, Friends: social.Friends, Blocks: social.Blocks, PendingInbox: social.PendingInbox, PendingOutbox: social.PendingOutbox, FriendCnt: len(social.Friends), BlockCnt: len(social.Blocks), PendingInboxCount: len(social.PendingInbox), PendingOutboxCount: len(social.PendingOutbox), RelationshipCount: len(social.RelationshipDetails), RelationshipDetails: social.RelationshipDetails}
	if s.parties != nil { party, appErr := s.parties.GetPartyByPlayer(ctx, playerID); if appErr != nil && appErr.Code != "not_found" { return PlayerOverview{}, appErr }; if appErr == nil { overview.CurrentPartyID = party.PartyID; if party.Queue != nil { overview.CurrentQueueStatus = party.Queue.Status; overview.CurrentQueueExpiresAt = party.Queue.ExpiresAt } } }
	if s.guilds != nil { guild, appErr := s.guilds.GetGuildByPlayer(ctx, playerID); if appErr != nil && appErr.Code != "not_found" { return PlayerOverview{}, appErr }; if appErr == nil { overview.CurrentGuildID = guild.GuildID; for _, member := range guild.Members { if member.PlayerID == playerID { overview.CurrentGuildRole = member.Role; break } } } }
	return overview, nil
}

func (s *OpsService) String() string { return fmt.Sprintf("ops-service(presence=%t,party=%t,guild=%t,worker=%t,social=%t,bootstrap=%t,redis=%t)", s.presence != nil, s.parties != nil, s.guilds != nil, s.worker != nil, s.social != nil, s.bootstrap != nil, s.redis != nil) }
