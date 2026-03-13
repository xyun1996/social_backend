package service

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/presence/internal/domain"
)

const (
	statusOnline  = "online"
	statusOffline = "offline"
)

// PresenceService provides an in-memory prototype for online state reporting.
type PresenceService struct {
	mu        sync.RWMutex
	presences map[string]domain.Presence
	now       func() time.Time
}

// NewPresenceService constructs an in-memory presence service.
func NewPresenceService() *PresenceService {
	return &PresenceService{
		presences: make(map[string]domain.Presence),
		now:       time.Now,
	}
}

// Connect marks a player online and records lightweight runtime metadata.
func (s *PresenceService) Connect(playerID string, sessionID string, realmID string, location string) (domain.Presence, *apperrors.Error) {
	if playerID == "" || sessionID == "" {
		err := apperrors.New("invalid_request", "player_id and session_id are required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	presence := domain.Presence{
		PlayerID:        playerID,
		Status:          statusOnline,
		SessionID:       sessionID,
		RealmID:         realmID,
		Location:        location,
		LastHeartbeatAt: now,
		LastSeenAt:      now,
		ConnectedAt:     &now,
	}

	s.presences[playerID] = presence
	return presence, nil
}

// Heartbeat refreshes the runtime metadata for an online player.
func (s *PresenceService) Heartbeat(playerID string, sessionID string, realmID string, location string) (domain.Presence, *apperrors.Error) {
	if playerID == "" || sessionID == "" {
		err := apperrors.New("invalid_request", "player_id and session_id are required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	presence, ok := s.presences[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", http.StatusNotFound)
		return domain.Presence{}, &err
	}

	if presence.SessionID != sessionID {
		err := apperrors.New("forbidden", "session_id does not match active presence", http.StatusForbidden)
		return domain.Presence{}, &err
	}

	now := s.now()
	presence.Status = statusOnline
	presence.RealmID = realmID
	presence.Location = location
	presence.LastHeartbeatAt = now
	presence.LastSeenAt = now
	presence.DisconnectedAt = nil
	s.presences[playerID] = presence
	return presence, nil
}

// Disconnect marks a player offline for the active session.
func (s *PresenceService) Disconnect(playerID string, sessionID string) (domain.Presence, *apperrors.Error) {
	if playerID == "" || sessionID == "" {
		err := apperrors.New("invalid_request", "player_id and session_id are required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	presence, ok := s.presences[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", http.StatusNotFound)
		return domain.Presence{}, &err
	}

	if presence.SessionID != sessionID {
		err := apperrors.New("forbidden", "session_id does not match active presence", http.StatusForbidden)
		return domain.Presence{}, &err
	}

	now := s.now()
	presence.Status = statusOffline
	presence.LastSeenAt = now
	presence.DisconnectedAt = &now
	s.presences[playerID] = presence
	return presence, nil
}

// GetPresence returns the current presence record for a player.
func (s *PresenceService) GetPresence(playerID string) (domain.Presence, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	presence, ok := s.presences[playerID]
	if !ok {
		err := apperrors.New("not_found", "presence not found", http.StatusNotFound)
		return domain.Presence{}, &err
	}

	return presence, nil
}

func (s *PresenceService) String() string {
	return fmt.Sprintf("presence-service(players=%d)", len(s.presences))
}
