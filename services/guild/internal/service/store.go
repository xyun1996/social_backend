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
}

type memoryGuildStore struct {
	mu         sync.RWMutex
	guilds     map[string]domain.Guild
	logs       map[string][]domain.GuildLogEntry
	activities map[string][]domain.GuildActivityRecord
}

func newMemoryGuildStore() *memoryGuildStore {
	return &memoryGuildStore{
		guilds:     make(map[string]domain.Guild),
		logs:       make(map[string][]domain.GuildLogEntry),
		activities: make(map[string][]domain.GuildActivityRecord),
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
