package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/presence/internal/domain"
)

const presenceKeyPrefix = "presence:player:"

// Store is the Redis foundation for future presence persistence.
type Store struct {
	config db.RedisConfig
	client *redis.Client
}

// NewStore constructs the presence Redis repository foundation.
func NewStore(config db.RedisConfig, client *redis.Client) *Store {
	return &Store{config: config, client: client}
}

// Key builds the canonical Redis key for a player presence snapshot.
func (s *Store) Key(playerID string) string {
	return presenceKeyPrefix + playerID
}

// Marshal encodes a presence snapshot for Redis storage.
func (s *Store) Marshal(presence domain.Presence) ([]byte, error) {
	return json.Marshal(presence)
}

// Unmarshal decodes a presence snapshot from Redis storage.
func (s *Store) Unmarshal(raw []byte) (domain.Presence, error) {
	var presence domain.Presence
	err := json.Unmarshal(raw, &presence)
	return presence, err
}

// URL returns the shared Redis URL used by this store.
func (s *Store) URL() string {
	return s.config.URL()
}

func (s *Store) String() string {
	return fmt.Sprintf("presence-redis-store(%s)", s.config.URL())
}

// SavePresence persists a presence snapshot in Redis.
func (s *Store) SavePresence(presence domain.Presence) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("redis store is not configured")
	}

	raw, err := s.Marshal(presence)
	if err != nil {
		return err
	}

	ttl := 24 * time.Hour
	if presence.Status == "online" {
		ttl = 2 * time.Hour
	}

	return s.client.Set(context.Background(), s.Key(presence.PlayerID), raw, ttl).Err()
}

// GetPresence loads the latest presence snapshot from Redis.
func (s *Store) GetPresence(playerID string) (domain.Presence, bool, error) {
	if s == nil || s.client == nil {
		return domain.Presence{}, false, fmt.Errorf("redis store is not configured")
	}

	raw, err := s.client.Get(context.Background(), s.Key(playerID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return domain.Presence{}, false, nil
		}
		return domain.Presence{}, false, err
	}

	presence, err := s.Unmarshal(raw)
	if err != nil {
		return domain.Presence{}, false, err
	}
	return presence, true, nil
}
