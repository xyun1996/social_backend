package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/xyun1996/social_backend/services/guild/internal/domain"
)

func (r *Repository) GetActivityByIdempotencyKey(guildID string, instanceID string, playerID string, idempotencyKey string) (domain.GuildActivityRecord, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.GuildActivityRecord{}, false, errors.New("mysql repository is not configured")
	}
	row := r.sqlDB.QueryRowContext(context.Background(), `SELECT activity_id, instance_id, guild_id, template_key, player_id, delta_xp, idempotency_key, source_type, created_at FROM guild_activities WHERE guild_id = ? AND instance_id = ? AND player_id = ? AND idempotency_key = ? LIMIT 1`, guildID, instanceID, playerID, idempotencyKey)
	var record domain.GuildActivityRecord
	if err := row.Scan(&record.ID, &record.InstanceID, &record.GuildID, &record.TemplateKey, &record.PlayerID, &record.DeltaXP, &record.IdempotencyKey, &record.SourceType, &record.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.GuildActivityRecord{}, false, nil
		}
		return domain.GuildActivityRecord{}, false, err
	}
	return record, true, nil
}

func (r *Repository) SaveContribution(record domain.GuildContribution) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO guild_contributions (guild_id, player_id, total_xp, last_source_type, updated_at) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE total_xp = VALUES(total_xp), last_source_type = VALUES(last_source_type), updated_at = VALUES(updated_at)`, record.GuildID, record.PlayerID, record.TotalXP, record.LastSourceType, record.UpdatedAt.UTC())
	return err
}

func (r *Repository) ListContributions(guildID string) ([]domain.GuildContribution, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT guild_id, player_id, total_xp, last_source_type, updated_at FROM guild_contributions WHERE guild_id = ?`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]domain.GuildContribution, 0)
	for rows.Next() {
		var record domain.GuildContribution
		if err := rows.Scan(&record.GuildID, &record.PlayerID, &record.TotalXP, &record.LastSourceType, &record.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.SortFunc(records, func(a domain.GuildContribution, b domain.GuildContribution) int {
		if a.TotalXP != b.TotalXP {
			if a.TotalXP > b.TotalXP {
				return -1
			}
			return 1
		}
		if a.PlayerID < b.PlayerID {
			return -1
		}
		if a.PlayerID > b.PlayerID {
			return 1
		}
		return 0
	})
	return records, nil
}

func (r *Repository) GetContribution(guildID string, playerID string) (domain.GuildContribution, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.GuildContribution{}, false, errors.New("mysql repository is not configured")
	}
	row := r.sqlDB.QueryRowContext(context.Background(), `SELECT guild_id, player_id, total_xp, last_source_type, updated_at FROM guild_contributions WHERE guild_id = ? AND player_id = ?`, guildID, playerID)
	var record domain.GuildContribution
	if err := row.Scan(&record.GuildID, &record.PlayerID, &record.TotalXP, &record.LastSourceType, &record.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.GuildContribution{}, false, nil
		}
		return domain.GuildContribution{}, false, err
	}
	return record, true, nil
}

func (r *Repository) SaveActivityInstance(record domain.GuildActivityInstance) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO guild_activity_instances (instance_id, guild_id, template_key, period_key, starts_at, ends_at, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE starts_at = VALUES(starts_at), ends_at = VALUES(ends_at), status = VALUES(status), updated_at = VALUES(updated_at)`, record.ID, record.GuildID, record.TemplateKey, record.PeriodKey, record.StartsAt.UTC(), record.EndsAt.UTC(), record.Status, record.CreatedAt.UTC(), record.UpdatedAt.UTC())
	return err
}

func (r *Repository) GetActivityInstance(guildID string, templateKey string, periodKey string) (domain.GuildActivityInstance, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.GuildActivityInstance{}, false, errors.New("mysql repository is not configured")
	}
	row := r.sqlDB.QueryRowContext(context.Background(), `SELECT instance_id, guild_id, template_key, period_key, starts_at, ends_at, status, created_at, updated_at FROM guild_activity_instances WHERE guild_id = ? AND template_key = ? AND period_key = ?`, guildID, templateKey, periodKey)
	var record domain.GuildActivityInstance
	if err := row.Scan(&record.ID, &record.GuildID, &record.TemplateKey, &record.PeriodKey, &record.StartsAt, &record.EndsAt, &record.Status, &record.CreatedAt, &record.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.GuildActivityInstance{}, false, nil
		}
		return domain.GuildActivityInstance{}, false, err
	}
	return record, true, nil
}

func (r *Repository) ListActivityInstances(guildID string, templateKey string) ([]domain.GuildActivityInstance, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}
	query := `SELECT instance_id, guild_id, template_key, period_key, starts_at, ends_at, status, created_at, updated_at FROM guild_activity_instances WHERE guild_id = ?`
	args := []any{guildID}
	if templateKey != "" {
		query += ` AND template_key = ?`
		args = append(args, templateKey)
	}
	rows, err := r.sqlDB.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]domain.GuildActivityInstance, 0)
	for rows.Next() {
		var record domain.GuildActivityInstance
		if err := rows.Scan(&record.ID, &record.GuildID, &record.TemplateKey, &record.PeriodKey, &record.StartsAt, &record.EndsAt, &record.Status, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.SortFunc(records, func(a domain.GuildActivityInstance, b domain.GuildActivityInstance) int {
		if !a.StartsAt.Equal(b.StartsAt) {
			if a.StartsAt.Before(b.StartsAt) {
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
	return records, nil
}

func (r *Repository) SaveRewardRecord(record domain.GuildRewardRecord) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO guild_reward_records (reward_id, guild_id, player_id, activity_id, template_key, reward_type, reward_ref, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE reward_type = VALUES(reward_type), reward_ref = VALUES(reward_ref), created_at = VALUES(created_at)`, record.ID, record.GuildID, record.PlayerID, record.ActivityID, record.TemplateKey, record.RewardType, record.RewardRef, record.CreatedAt.UTC())
	return err
}

func (r *Repository) ListRewardRecords(guildID string) ([]domain.GuildRewardRecord, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT reward_id, guild_id, player_id, activity_id, template_key, reward_type, reward_ref, created_at FROM guild_reward_records WHERE guild_id = ?`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]domain.GuildRewardRecord, 0)
	for rows.Next() {
		var record domain.GuildRewardRecord
		if err := rows.Scan(&record.ID, &record.GuildID, &record.PlayerID, &record.ActivityID, &record.TemplateKey, &record.RewardType, &record.RewardRef, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.SortFunc(records, func(a domain.GuildRewardRecord, b domain.GuildRewardRecord) int {
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
	return records, nil
}
