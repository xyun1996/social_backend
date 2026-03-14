package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/guild/internal/domain"
)

// GetProgression returns the current progression read model.
func (s *GuildService) GetProgression(guildID string) (domain.GuildProgression, *apperrors.Error) {
	guild, appErr := s.GetGuild(guildID)
	if appErr != nil {
		return domain.GuildProgression{}, appErr
	}
	return progressionForGuild(guild, s.now().UTC()), nil
}

// ListContributions returns the current contribution leaderboard.
func (s *GuildService) ListContributions(guildID string) ([]domain.GuildContribution, *apperrors.Error) {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return nil, &err
	}
	if _, ok, err := s.guilds.GetGuild(guildID); err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	} else if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return nil, &err
	}
	records, err := s.guilds.ListContributions(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	return records, nil
}

// ListActivityInstances returns the currently known instances for a template.
func (s *GuildService) ListActivityInstances(guildID string, templateKey string) ([]domain.GuildActivityInstance, *apperrors.Error) {
	if guildID == "" || templateKey == "" {
		err := apperrors.New("invalid_request", "guild_id and template_key are required", http.StatusBadRequest)
		return nil, &err
	}
	if _, ok, err := s.guilds.GetGuild(guildID); err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	} else if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return nil, &err
	}
	instances, err := s.guilds.ListActivityInstances(guildID, templateKey)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	return instances, nil
}

// ListRewardRecords returns reward bookkeeping records for a guild.
func (s *GuildService) ListRewardRecords(guildID string) ([]domain.GuildRewardRecord, *apperrors.Error) {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return nil, &err
	}
	if _, ok, err := s.guilds.GetGuild(guildID); err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	} else if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return nil, &err
	}
	records, err := s.guilds.ListRewardRecords(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	return records, nil
}

// SubmitActivityWithOptions applies a fixed activity template with idempotency and source metadata.
func (s *GuildService) SubmitActivityWithOptions(ctx context.Context, guildID string, actorPlayerID string, templateKey string, idempotencyKey string, sourceType string) (domain.GuildActivityRecord, domain.Guild, domain.GuildProgression, *apperrors.Error) {
	if guildID == "" || actorPlayerID == "" || templateKey == "" {
		err := apperrors.New("invalid_request", "guild_id, actor_player_id, and template_key are required", http.StatusBadRequest)
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &err
	}

	guild, ok, err := s.guilds.GetGuild(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "guild not found", http.StatusNotFound)
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &err
	}
	if !hasMember(guild.Members, actorPlayerID) {
		err := apperrors.New("forbidden", "only guild members can submit activities", http.StatusForbidden)
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &err
	}

	template, appErr := s.lookupTemplate(templateKey)
	if appErr != nil {
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, appErr
	}

	now := s.now().UTC()
	if appErr := s.CloseExpiredActivityInstances(guildID, now); appErr != nil {
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, appErr
	}
	instance, created, appErr := s.ensureActivityInstance(guildID, template, now)
	if appErr != nil {
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, appErr
	}
	if created {
		if appErr := s.appendLog(guild.ID, actionGuildActivityOpened, systemSenderID, "", "guild activity opened: "+template.Key); appErr != nil {
			return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, appErr
		}
		_ = s.publishGuildEvent(ctx, guild, fmt.Sprintf("Guild activity %s is now active.", template.Name))
		_ = s.enqueueActivityLifecycleJobs(ctx, guild.ID, instance)
	}

	if idempotencyKey != "" {
		existing, ok, err := s.guilds.GetActivityByIdempotencyKey(guildID, instance.ID, actorPlayerID, idempotencyKey)
		if err != nil {
			internal := apperrors.Internal()
			return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
		}
		if ok {
			return existing, guild, progressionForGuild(guild, now), nil
		}
	}

	records, err := s.guilds.ListActivities(guildID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
	}
	submissionCount := 0
	for _, record := range records {
		if record.InstanceID == instance.ID && record.PlayerID == actorPlayerID {
			submissionCount++
		}
	}
	if submissionCount >= template.MaxSubmissionsPerPeriod {
		err := apperrors.New("submission_limit_reached", "activity submission limit reached for the current period", http.StatusConflict)
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &err
	}

	activityID, idErr := s.newActivityID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
	}
	record := domain.GuildActivityRecord{
		ID:             activityID,
		InstanceID:     instance.ID,
		GuildID:        guildID,
		TemplateKey:    template.Key,
		PlayerID:       actorPlayerID,
		DeltaXP:        template.ContributionXP,
		IdempotencyKey: idempotencyKey,
		SourceType:     sourceType,
		CreatedAt:      now,
	}

	contribution, exists, err := s.guilds.GetContribution(guildID, actorPlayerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
	}
	if !exists {
		contribution = domain.GuildContribution{GuildID: guildID, PlayerID: actorPlayerID}
	}
	contribution.TotalXP += template.ContributionXP
	contribution.LastSourceType = template.Key
	contribution.UpdatedAt = now

	beforeLevel := guild.Level
	guild.Experience += template.ContributionXP
	guild.Level = max(1, 1+(guild.Experience/guildXPPerLevel))
	if err := s.guilds.SaveGuild(guild); err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
	}
	if err := s.guilds.SaveActivity(record); err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
	}
	if err := s.guilds.SaveContribution(contribution); err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
	}
	if template.RewardType != "" {
		rewardID, idErr := s.newRewardID()
		if idErr != nil {
			internal := apperrors.Internal()
			return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
		}
		reward := domain.GuildRewardRecord{ID: rewardID, GuildID: guildID, PlayerID: actorPlayerID, ActivityID: record.ID, TemplateKey: template.Key, RewardType: template.RewardType, RewardRef: template.RewardRef, CreatedAt: now}
		if err := s.guilds.SaveRewardRecord(reward); err != nil {
			internal := apperrors.Internal()
			return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, &internal
		}
	}
	if appErr := s.appendLog(guild.ID, actionGuildActivitySubmitted, actorPlayerID, "", "guild activity submitted: "+template.Key); appErr != nil {
		return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, appErr
	}
	_ = s.publishGuildEvent(ctx, guild, fmt.Sprintf("%s completed %s for +%d guild xp.", actorPlayerID, template.Name, template.ContributionXP))
	if guild.Level > beforeLevel {
		if appErr := s.appendLog(guild.ID, actionGuildLeveledUp, actorPlayerID, "", fmt.Sprintf("guild reached level %d", guild.Level)); appErr != nil {
			return domain.GuildActivityRecord{}, domain.Guild{}, domain.GuildProgression{}, appErr
		}
		_ = s.publishGuildEvent(ctx, guild, fmt.Sprintf("Guild leveled up to %d.", guild.Level))
	}
	return record, guild, progressionForGuild(guild, now), nil
}

// EnsureCurrentActivityInstances guarantees current period instances exist for every template.
func (s *GuildService) EnsureCurrentActivityInstances(guildID string, at time.Time) ([]domain.GuildActivityInstance, *apperrors.Error) {
	guild, appErr := s.GetGuild(guildID)
	if appErr != nil {
		return nil, appErr
	}
	at = at.UTC()
	instances := make([]domain.GuildActivityInstance, 0, len(s.ListActivityTemplates()))
	for _, template := range s.ListActivityTemplates() {
		instance, created, appErr := s.ensureActivityInstance(guild.ID, template, at)
		if appErr != nil {
			return nil, appErr
		}
		if created {
			if appErr := s.appendLog(guild.ID, actionGuildActivityOpened, systemSenderID, "", "guild activity opened: "+template.Key); appErr != nil {
				return nil, appErr
			}
			_ = s.publishGuildEvent(context.Background(), guild, fmt.Sprintf("Guild activity %s is now active.", template.Name))
			_ = s.enqueueActivityLifecycleJobs(context.Background(), guild.ID, instance)
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

// CloseExpiredActivityInstances closes active instances whose period already ended.
func (s *GuildService) CloseExpiredActivityInstances(guildID string, at time.Time) *apperrors.Error {
	if guildID == "" {
		err := apperrors.New("invalid_request", "guild_id is required", http.StatusBadRequest)
		return &err
	}
	instances, err := s.guilds.ListActivityInstances(guildID, "")
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	for _, instance := range instances {
		if instance.Status != activityStatusActive {
			continue
		}
		if instance.EndsAt.After(at.UTC()) {
			continue
		}
		instance.Status = activityStatusClosed
		instance.UpdatedAt = at.UTC()
		if err := s.guilds.SaveActivityInstance(instance); err != nil {
			internal := apperrors.Internal()
			return &internal
		}
	}
	return nil
}

func (s *GuildService) lookupTemplate(templateKey string) (domain.GuildActivityTemplate, *apperrors.Error) {
	for _, candidate := range s.ListActivityTemplates() {
		if candidate.Key == templateKey {
			return candidate, nil
		}
	}
	err := apperrors.New("not_found", "activity template not found", http.StatusNotFound)
	return domain.GuildActivityTemplate{}, &err
}

func (s *GuildService) ensureActivityInstance(guildID string, template domain.GuildActivityTemplate, at time.Time) (domain.GuildActivityInstance, bool, *apperrors.Error) {
	periodKey, startsAt, endsAt := activityWindow(template.PeriodType, at)
	instance, ok, err := s.guilds.GetActivityInstance(guildID, template.Key, periodKey)
	if err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityInstance{}, false, &internal
	}
	if ok {
		if instance.Status != activityStatusActive && endsAt.After(at) {
			instance.Status = activityStatusActive
			instance.UpdatedAt = at
			if err := s.guilds.SaveActivityInstance(instance); err != nil {
				internal := apperrors.Internal()
				return domain.GuildActivityInstance{}, false, &internal
			}
		}
		return instance, false, nil
	}
	instanceID, idErr := s.newInstanceID()
	if idErr != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityInstance{}, false, &internal
	}
	instance = domain.GuildActivityInstance{ID: instanceID, GuildID: guildID, TemplateKey: template.Key, PeriodKey: periodKey, StartsAt: startsAt, EndsAt: endsAt, Status: activityStatusActive, CreatedAt: at, UpdatedAt: at}
	if err := s.guilds.SaveActivityInstance(instance); err != nil {
		internal := apperrors.Internal()
		return domain.GuildActivityInstance{}, false, &internal
	}
	return instance, true, nil
}

func activityWindow(periodType string, at time.Time) (string, time.Time, time.Time) {
	utc := at.UTC()
	switch periodType {
	case "weekly":
		weekday := int(utc.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
		end := start.AddDate(0, 0, 7)
		isoYear, isoWeek := start.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", isoYear, isoWeek), start, end
	default:
		start := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 1)
		return start.Format("2006-01-02"), start, end
	}
}

func progressionForGuild(guild domain.Guild, now time.Time) domain.GuildProgression {
	nextLevelXP := guild.Level * guildXPPerLevel
	if nextLevelXP <= guild.Experience {
		nextLevelXP = (guild.Level + 1) * guildXPPerLevel
	}
	return domain.GuildProgression{GuildID: guild.ID, Level: guild.Level, Experience: guild.Experience, NextLevelXP: nextLevelXP, UpdatedAt: now}
}

func (s *GuildService) publishGuildEvent(ctx context.Context, guild domain.Guild, body string) *apperrors.Error {
	if s.chat == nil {
		return nil
	}
	memberIDs := make([]string, 0, len(guild.Members))
	for _, member := range guild.Members {
		memberIDs = append(memberIDs, member.PlayerID)
	}
	return s.chat.PublishGuildSystemEvent(ctx, guild.ID, memberIDs, body)
}

func (s *GuildService) enqueueActivityLifecycleJobs(ctx context.Context, guildID string, instance domain.GuildActivityInstance) *apperrors.Error {
	if s.scheduler == nil {
		return nil
	}
	payload := fmt.Sprintf(`{"guild_id":"%s","template_key":"%s","period_key":"%s","instance_id":"%s"}`, guildID, instance.TemplateKey, instance.PeriodKey, instance.ID)
	if appErr := s.scheduler.EnqueueJob(ctx, guildActivityEnsureJobType, payload); appErr != nil {
		return appErr
	}
	return s.scheduler.EnqueueJob(ctx, guildActivityCloseJobType, payload)
}
