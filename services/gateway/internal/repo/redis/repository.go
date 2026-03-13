package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

const (
	sessionsIndexKey    = "gateway:sessions"
	sessionKeyPrefix    = "gateway:session:"
	sessionEventsPrefix = "gateway:session_events:"
)

// Repository persists gateway realtime session state in Redis.
type Repository struct {
	config db.RedisConfig
	client *redis.Client
}

func NewRepository(config db.RedisConfig, client *redis.Client) *Repository {
	return &Repository{config: config, client: client}
}

func (r *Repository) URL() string {
	return r.config.URL()
}

func (r *Repository) SessionKey(sessionID string) string {
	return sessionKeyPrefix + sessionID
}

func (r *Repository) EventsKey(sessionID string) string {
	return sessionEventsPrefix + sessionID
}

func (r *Repository) SaveSession(session gatewayservice.RealtimeSession) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis repository is not configured")
	}
	raw, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ctx := context.Background()
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, r.SessionKey(session.SessionID), raw, 0)
	pipe.SAdd(ctx, sessionsIndexKey, session.SessionID)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *Repository) GetSession(sessionID string) (gatewayservice.RealtimeSession, bool, error) {
	if r == nil || r.client == nil {
		return gatewayservice.RealtimeSession{}, false, fmt.Errorf("redis repository is not configured")
	}

	raw, err := r.client.Get(context.Background(), r.SessionKey(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return gatewayservice.RealtimeSession{}, false, nil
		}
		return gatewayservice.RealtimeSession{}, false, err
	}

	var session gatewayservice.RealtimeSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return gatewayservice.RealtimeSession{}, false, err
	}
	return session, true, nil
}

func (r *Repository) ListSessions() ([]gatewayservice.RealtimeSession, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("redis repository is not configured")
	}

	ids, err := r.client.SMembers(context.Background(), sessionsIndexKey).Result()
	if err != nil {
		return nil, err
	}
	sort.Strings(ids)

	sessions := make([]gatewayservice.RealtimeSession, 0, len(ids))
	for _, id := range ids {
		session, ok, err := r.GetSession(id)
		if err != nil {
			return nil, err
		}
		if ok {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (r *Repository) SaveEvents(sessionID string, events []gatewayservice.ChatMessageEnvelope) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis repository is not configured")
	}
	raw, err := json.Marshal(events)
	if err != nil {
		return err
	}
	return r.client.Set(context.Background(), r.EventsKey(sessionID), raw, 0).Err()
}

func (r *Repository) GetEvents(sessionID string) ([]gatewayservice.ChatMessageEnvelope, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("redis repository is not configured")
	}
	raw, err := r.client.Get(context.Background(), r.EventsKey(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return []gatewayservice.ChatMessageEnvelope{}, nil
		}
		return nil, err
	}
	var events []gatewayservice.ChatMessageEnvelope
	if err := json.Unmarshal(raw, &events); err != nil {
		return nil, err
	}
	return events, nil
}
