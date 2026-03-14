package service

import (
	"slices"
	"sync"

	"github.com/xyun1996/social_backend/services/guild/internal/domain"
)

// GuildStore persists guild aggregate state.
type GuildStore interface {
	SaveGuild(guild domain.Guild) error
	GetGuild(guildID string) (domain.Guild, bool, error)
	ListGuilds() ([]domain.Guild, error)
	SaveLog(entry domain.GuildLogEntry) error
	ListLogs(guildID string) ([]domain.GuildLogEntry, error)
	SaveActivity(record domain.GuildActivityRecord) error
	ListActivities(guildID string) ([]domain.GuildActivityRecord, error)
	GetActivityByIdempotencyKey(guildID string, instanceID string, playerID string, idempotencyKey string) (domain.GuildActivityRecord, bool, error)
	SaveContribution(record domain.GuildContribution) error
	ListContributions(guildID string) ([]domain.GuildContribution, error)
	GetContribution(guildID string, playerID string) (domain.GuildContribution, bool, error)
	SaveActivityInstance(record domain.GuildActivityInstance) error
	GetActivityInstance(guildID string, templateKey string, periodKey string) (domain.GuildActivityInstance, bool, error)
	ListActivityInstances(guildID string, templateKey string) ([]domain.GuildActivityInstance, error)
	SaveRewardRecord(record domain.GuildRewardRecord) error
	ListRewardRecords(guildID string) ([]domain.GuildRewardRecord, error)
}

type memoryGuildStore struct {
	mu            sync.RWMutex
	guilds        map[string]domain.Guild
	logs          map[string][]domain.GuildLogEntry
	activities    map[string][]domain.GuildActivityRecord
	contributions map[string]map[string]domain.GuildContribution
	instances     map[string]map[string]domain.GuildActivityInstance
	rewards       map[string][]domain.GuildRewardRecord
}

func newMemoryGuildStore() *memoryGuildStore {
	return &memoryGuildStore{
		guilds:        make(map[string]domain.Guild),
		logs:          make(map[string][]domain.GuildLogEntry),
		activities:    make(map[string][]domain.GuildActivityRecord),
		contributions: make(map[string]map[string]domain.GuildContribution),
		instances:     make(map[string]map[string]domain.GuildActivityInstance),
		rewards:       make(map[string][]domain.GuildRewardRecord),
	}
}

func (s *memoryGuildStore) SaveGuild(guild domain.Guild) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.guilds[guild.ID] = guild
	return nil
}

func (s *memoryGuildStore) GetGuild(guildID string) (domain.Guild, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	guild, ok := s.guilds[guildID]
	return guild, ok, nil
}

func (s *memoryGuildStore) ListGuilds() ([]domain.Guild, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guilds := make([]domain.Guild, 0, len(s.guilds))
	for _, guild := range s.guilds {
		guilds = append(guilds, guild)
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

func (s *memoryGuildStore) SaveLog(entry domain.GuildLogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs[entry.GuildID] = append(s.logs[entry.GuildID], entry)
	return nil
}

func (s *memoryGuildStore) ListLogs(guildID string) ([]domain.GuildLogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := append([]domain.GuildLogEntry(nil), s.logs[guildID]...)
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

func (s *memoryGuildStore) SaveActivity(record domain.GuildActivityRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activities[record.GuildID] = append(s.activities[record.GuildID], record)
	return nil
}

func (s *memoryGuildStore) ListActivities(guildID string) ([]domain.GuildActivityRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := append([]domain.GuildActivityRecord(nil), s.activities[guildID]...)
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

func (s *memoryGuildStore) GetActivityByIdempotencyKey(guildID string, instanceID string, playerID string, idempotencyKey string) (domain.GuildActivityRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, record := range s.activities[guildID] {
		if record.InstanceID == instanceID && record.PlayerID == playerID && record.IdempotencyKey == idempotencyKey {
			return record, true, nil
		}
	}
	return domain.GuildActivityRecord{}, false, nil
}

func (s *memoryGuildStore) SaveContribution(record domain.GuildContribution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.contributions[record.GuildID] == nil {
		s.contributions[record.GuildID] = make(map[string]domain.GuildContribution)
	}
	s.contributions[record.GuildID][record.PlayerID] = record
	return nil
}

func (s *memoryGuildStore) ListContributions(guildID string) ([]domain.GuildContribution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	guildContributions := s.contributions[guildID]
	records := make([]domain.GuildContribution, 0, len(guildContributions))
	for _, record := range guildContributions {
		records = append(records, record)
	}
	slices.SortFunc(records, func(a domain.GuildContribution, b domain.GuildContribution) int {
		if a.TotalXP != b.TotalXP {
			if a.TotalXP > b.TotalXP {
				return -1
			}
			return 1
		}
		switch {
		case a.PlayerID < b.PlayerID:
			return -1
		case a.PlayerID > b.PlayerID:
			return 1
		default:
			return 0
		}
	})
	return records, nil
}

func (s *memoryGuildStore) GetContribution(guildID string, playerID string) (domain.GuildContribution, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.contributions[guildID][playerID]
	return record, ok, nil
}

func (s *memoryGuildStore) SaveActivityInstance(record domain.GuildActivityInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.instances[record.GuildID] == nil {
		s.instances[record.GuildID] = make(map[string]domain.GuildActivityInstance)
	}
	s.instances[record.GuildID][record.TemplateKey+":"+record.PeriodKey] = record
	return nil
}

func (s *memoryGuildStore) GetActivityInstance(guildID string, templateKey string, periodKey string) (domain.GuildActivityInstance, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.instances[guildID][templateKey+":"+periodKey]
	return record, ok, nil
}

func (s *memoryGuildStore) ListActivityInstances(guildID string, templateKey string) ([]domain.GuildActivityInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	instanceMap := s.instances[guildID]
	records := make([]domain.GuildActivityInstance, 0, len(instanceMap))
	for _, record := range instanceMap {
		if templateKey != "" && record.TemplateKey != templateKey {
			continue
		}
		records = append(records, record)
	}
	slices.SortFunc(records, func(a domain.GuildActivityInstance, b domain.GuildActivityInstance) int {
		if !a.StartsAt.Equal(b.StartsAt) {
			if a.StartsAt.Before(b.StartsAt) {
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

func (s *memoryGuildStore) SaveRewardRecord(record domain.GuildRewardRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rewards[record.GuildID] = append(s.rewards[record.GuildID], record)
	return nil
}

func (s *memoryGuildStore) ListRewardRecords(guildID string) ([]domain.GuildRewardRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := append([]domain.GuildRewardRecord(nil), s.rewards[guildID]...)
	slices.SortFunc(records, func(a domain.GuildRewardRecord, b domain.GuildRewardRecord) int {
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
