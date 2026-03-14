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
}

type memoryGuildStore struct {
	mu     sync.RWMutex
	guilds map[string]domain.Guild
	logs   map[string][]domain.GuildLogEntry
}

func newMemoryGuildStore() *memoryGuildStore {
	return &memoryGuildStore{
		guilds: make(map[string]domain.Guild),
		logs:   make(map[string][]domain.GuildLogEntry),
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
