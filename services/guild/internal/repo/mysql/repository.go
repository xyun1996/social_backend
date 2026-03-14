package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/guild/internal/domain"
)

type schemaExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const (
	GuildsTable               = "guild_guilds"
	GuildMembersTable         = "guild_members"
	GuildLogsTable            = "guild_logs"
	GuildActivitiesTable      = "guild_activities"
	GuildContributionsTable   = "guild_contributions"
	GuildActivityInstancesTbl = "guild_activity_instances"
	GuildRewardsTable         = "guild_reward_records"
)

// Repository implements guild durable storage on MySQL.
type Repository struct {
	config db.MySQLConfig
	sqlDB  *sql.DB
}

func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository {
	return &Repository{config: config, sqlDB: sqlDB}
}

func (r *Repository) DSN() string {
	return r.config.DSN()
}

// Migrations returns the versioned guild schema ownership.
func (r *Repository) Migrations() []db.Migration {
	return []db.Migration{
		{
			ID: "001_guild_core",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS guild_guilds (
					guild_id VARCHAR(64) PRIMARY KEY,
					name VARCHAR(128) NOT NULL,
					owner_id VARCHAR(64) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);`,
				`CREATE TABLE IF NOT EXISTS guild_members (
					guild_id VARCHAR(64) NOT NULL,
					player_id VARCHAR(64) NOT NULL,
					role VARCHAR(32) NOT NULL,
					joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (guild_id, player_id),
					INDEX idx_guild_members_player (player_id)
				);`,
			},
		},
		{
			ID: "002_guild_announcement",
			Statements: []string{
				`ALTER TABLE guild_guilds
				 ADD COLUMN announcement VARCHAR(2048) NOT NULL DEFAULT ''`,
				`ALTER TABLE guild_guilds
				 ADD COLUMN announcement_updated_at TIMESTAMP NULL DEFAULT NULL`,
			},
		},
		{
			ID: "003_guild_logs",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS guild_logs (
					log_id VARCHAR(64) PRIMARY KEY,
					guild_id VARCHAR(64) NOT NULL,
					action VARCHAR(64) NOT NULL,
					actor_id VARCHAR(64) NOT NULL DEFAULT '',
					target_id VARCHAR(64) NOT NULL DEFAULT '',
					message TEXT NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_guild_logs_guild_created (guild_id, created_at)
				);`,
			},
		},
		{
			ID: "004_guild_progression",
			Statements: []string{
				`ALTER TABLE guild_guilds
				 ADD COLUMN level INT NOT NULL DEFAULT 1`,
				`ALTER TABLE guild_guilds
				 ADD COLUMN experience INT NOT NULL DEFAULT 0`,
				`CREATE TABLE IF NOT EXISTS guild_activities (
					activity_id VARCHAR(64) PRIMARY KEY,
					guild_id VARCHAR(64) NOT NULL,
					template_key VARCHAR(64) NOT NULL,
					player_id VARCHAR(64) NOT NULL,
					delta_xp INT NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_guild_activities_guild_created (guild_id, created_at)
				);`,
			},
		},
		{
			ID: "005_guild_progression_v2",
			Statements: []string{
				`ALTER TABLE guild_activities
				 ADD COLUMN instance_id VARCHAR(64) NOT NULL DEFAULT ''`,
				`ALTER TABLE guild_activities
				 ADD COLUMN idempotency_key VARCHAR(128) NOT NULL DEFAULT ''`,
				`ALTER TABLE guild_activities
				 ADD COLUMN source_type VARCHAR(64) NOT NULL DEFAULT ''`,
				`CREATE TABLE IF NOT EXISTS guild_contributions (
					guild_id VARCHAR(64) NOT NULL,
					player_id VARCHAR(64) NOT NULL,
					total_xp INT NOT NULL,
					last_source_type VARCHAR(64) NOT NULL DEFAULT '',
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (guild_id, player_id),
					INDEX idx_guild_contributions_total (guild_id, total_xp)
				);`,
				`CREATE TABLE IF NOT EXISTS guild_activity_instances (
					instance_id VARCHAR(64) PRIMARY KEY,
					guild_id VARCHAR(64) NOT NULL,
					template_key VARCHAR(64) NOT NULL,
					period_key VARCHAR(64) NOT NULL,
					starts_at TIMESTAMP NOT NULL,
					ends_at TIMESTAMP NOT NULL,
					status VARCHAR(32) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					UNIQUE KEY uq_guild_activity_instance (guild_id, template_key, period_key)
				);`,
				`CREATE TABLE IF NOT EXISTS guild_reward_records (
					reward_id VARCHAR(64) PRIMARY KEY,
					guild_id VARCHAR(64) NOT NULL,
					player_id VARCHAR(64) NOT NULL,
					activity_id VARCHAR(64) NOT NULL,
					template_key VARCHAR(64) NOT NULL,
					reward_type VARCHAR(64) NOT NULL,
					reward_ref VARCHAR(128) NOT NULL DEFAULT '',
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_guild_reward_records_guild_created (guild_id, created_at)
				);`,
			},
		},
	}
}

func (r *Repository) SchemaStatements() []string {
	return db.FlattenMigrations(r.Migrations())
}

func (r *Repository) BootstrapSchema(ctx context.Context) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	return db.ApplyMySQLMigrations(ctx, r.sqlDB, "guild", r.Migrations())
}

func (r *Repository) SaveGuild(guild domain.Guild) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	tx, err := r.sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(
		context.Background(),
		`INSERT INTO guild_guilds (guild_id, name, owner_id, level, experience, announcement, announcement_updated_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   name = VALUES(name),
		   owner_id = VALUES(owner_id),
		   level = VALUES(level),
		   experience = VALUES(experience),
		   announcement = VALUES(announcement),
		   announcement_updated_at = VALUES(announcement_updated_at),
		   created_at = VALUES(created_at)`,
		guild.ID,
		guild.Name,
		guild.OwnerID,
		guild.Level,
		guild.Experience,
		guild.Announcement,
		nullableTime(guild.AnnouncementUpdatedAt),
		guild.CreatedAt.UTC(),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(), `DELETE FROM guild_members WHERE guild_id = ?`, guild.ID); err != nil {
		return err
	}
	for _, member := range guild.Members {
		if _, err := tx.ExecContext(
			context.Background(),
			`INSERT INTO guild_members (guild_id, player_id, role, joined_at) VALUES (?, ?, ?, ?)`,
			guild.ID,
			member.PlayerID,
			member.Role,
			member.JoinedAt.UTC(),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetGuild(guildID string) (domain.Guild, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.Guild{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT guild_id, name, owner_id, level, experience, announcement, announcement_updated_at, created_at FROM guild_guilds WHERE guild_id = ?`,
		guildID,
	)

	var guild domain.Guild
	var announcementUpdatedAt sql.NullTime
	if err := row.Scan(&guild.ID, &guild.Name, &guild.OwnerID, &guild.Level, &guild.Experience, &guild.Announcement, &announcementUpdatedAt, &guild.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Guild{}, false, nil
		}
		return domain.Guild{}, false, err
	}
	if announcementUpdatedAt.Valid {
		guild.AnnouncementUpdatedAt = announcementUpdatedAt.Time
	}

	members, err := r.listMembers(guildID)
	if err != nil {
		return domain.Guild{}, false, err
	}
	guild.Members = members
	return guild, true, nil
}

func (r *Repository) ListGuilds() ([]domain.Guild, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT guild_id, name, owner_id, level, experience, announcement, announcement_updated_at, created_at FROM guild_guilds`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	guilds := make([]domain.Guild, 0)
	for rows.Next() {
		var guild domain.Guild
		var announcementUpdatedAt sql.NullTime
		if err := rows.Scan(&guild.ID, &guild.Name, &guild.OwnerID, &guild.Level, &guild.Experience, &guild.Announcement, &announcementUpdatedAt, &guild.CreatedAt); err != nil {
			return nil, err
		}
		if announcementUpdatedAt.Valid {
			guild.AnnouncementUpdatedAt = announcementUpdatedAt.Time
		}
		members, err := r.listMembers(guild.ID)
		if err != nil {
			return nil, err
		}
		guild.Members = members
		guilds = append(guilds, guild)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(guilds, func(a domain.Guild, b domain.Guild) int {
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
	return guilds, nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}

func (r *Repository) listMembers(guildID string) ([]domain.GuildMember, error) {
	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT player_id, role, joined_at FROM guild_members WHERE guild_id = ?`,
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.GuildMember, 0)
	for rows.Next() {
		var member domain.GuildMember
		if err := rows.Scan(&member.PlayerID, &member.Role, &member.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(members, func(a domain.GuildMember, b domain.GuildMember) int {
		switch {
		case a.PlayerID < b.PlayerID:
			return -1
		case a.PlayerID > b.PlayerID:
			return 1
		default:
			return 0
		}
	})
	return members, nil
}

func (r *Repository) SaveLog(entry domain.GuildLogEntry) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO guild_logs (log_id, guild_id, action, actor_id, target_id, message, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   action = VALUES(action),
		   actor_id = VALUES(actor_id),
		   target_id = VALUES(target_id),
		   message = VALUES(message),
		   created_at = VALUES(created_at)`,
		entry.ID,
		entry.GuildID,
		entry.Action,
		entry.ActorID,
		entry.TargetID,
		entry.Message,
		entry.CreatedAt.UTC(),
	)
	return err
}

func (r *Repository) ListLogs(guildID string) ([]domain.GuildLogEntry, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT log_id, guild_id, action, actor_id, target_id, message, created_at
		 FROM guild_logs
		 WHERE guild_id = ?`,
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.GuildLogEntry, 0)
	for rows.Next() {
		var entry domain.GuildLogEntry
		if err := rows.Scan(&entry.ID, &entry.GuildID, &entry.Action, &entry.ActorID, &entry.TargetID, &entry.Message, &entry.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(logs, func(a domain.GuildLogEntry, b domain.GuildLogEntry) int {
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
	return logs, nil
}

func (r *Repository) SaveActivity(record domain.GuildActivityRecord) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO guild_activities (activity_id, instance_id, guild_id, template_key, player_id, delta_xp, idempotency_key, source_type, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   instance_id = VALUES(instance_id),
		   template_key = VALUES(template_key),
		   player_id = VALUES(player_id),
		   delta_xp = VALUES(delta_xp),
		   idempotency_key = VALUES(idempotency_key),
		   source_type = VALUES(source_type),
		   created_at = VALUES(created_at)`,
		record.ID,
		record.InstanceID,
		record.GuildID,
		record.TemplateKey,
		record.PlayerID,
		record.DeltaXP,
		record.IdempotencyKey,
		record.SourceType,
		record.CreatedAt.UTC(),
	)
	return err
}

func (r *Repository) ListActivities(guildID string) ([]domain.GuildActivityRecord, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}
	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT activity_id, instance_id, guild_id, template_key, player_id, delta_xp, idempotency_key, source_type, created_at
		 FROM guild_activities
		 WHERE guild_id = ?`,
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.GuildActivityRecord, 0)
	for rows.Next() {
		var record domain.GuildActivityRecord
		if err := rows.Scan(&record.ID, &record.InstanceID, &record.GuildID, &record.TemplateKey, &record.PlayerID, &record.DeltaXP, &record.IdempotencyKey, &record.SourceType, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.SortFunc(records, func(a domain.GuildActivityRecord, b domain.GuildActivityRecord) int {
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
	return records, nil
}
