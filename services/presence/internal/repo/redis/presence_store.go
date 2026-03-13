package redis

import (
	"encoding/json"
	"fmt"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/presence/internal/domain"
)

const presenceKeyPrefix = "presence:player:"

// Store is the Redis foundation for future presence persistence.
type Store struct {
	config db.RedisConfig
}

// NewStore constructs the presence Redis repository foundation.
func NewStore(config db.RedisConfig) *Store {
	return &Store{config: config}
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
