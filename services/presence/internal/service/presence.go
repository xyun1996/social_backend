package service

import (
	"fmt"
	"net/http"
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
	store PresenceStore
	now   func() time.Time
}

// NewPresenceService constructs an in-memory presence service.
func NewPresenceService() *PresenceService {
	return &PresenceService{
		store: newMemoryPresenceStore(),
		now:   time.Now,
	}
}

// NewPresenceServiceWithStore constructs a presence service with a custom store.
func NewPresenceServiceWithStore(store PresenceStore) *PresenceService {
	if store == nil {
		return NewPresenceService()
	}

	return &PresenceService{
		store: store,
		now:   time.Now,
	}
}

// Connect marks a player online and records lightweight runtime metadata.
func (s *PresenceService) Connect(playerID string, sessionID string, realmID string, location string) (domain.Presence, *apperrors.Error) {
	if playerID == "" || sessionID == "" {
		err := apperrors.New("invalid_request", "player_id and session_id are required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

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

	if err := s.store.SavePresence(presence); err != nil {
		internal := apperrors.Internal()
		return domain.Presence{}, &internal
	}
	return presence, nil
}

// Heartbeat refreshes the runtime metadata for an online player.
func (s *PresenceService) Heartbeat(playerID string, sessionID string, realmID string, location string) (domain.Presence, *apperrors.Error) {
	if playerID == "" || sessionID == "" {
		err := apperrors.New("invalid_request", "player_id and session_id are required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

	presence, ok, err := s.store.GetPresence(playerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Presence{}, &internal
	}
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
	if err := s.store.SavePresence(presence); err != nil {
		internal := apperrors.Internal()
		return domain.Presence{}, &internal
	}
	return presence, nil
}

// Disconnect marks a player offline for the active session.
func (s *PresenceService) Disconnect(playerID string, sessionID string) (domain.Presence, *apperrors.Error) {
	if playerID == "" || sessionID == "" {
		err := apperrors.New("invalid_request", "player_id and session_id are required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

	presence, ok, err := s.store.GetPresence(playerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Presence{}, &internal
	}
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
	if err := s.store.SavePresence(presence); err != nil {
		internal := apperrors.Internal()
		return domain.Presence{}, &internal
	}
	return presence, nil
}

// GetPresence returns the current presence record for a player.
func (s *PresenceService) GetPresence(playerID string) (domain.Presence, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return domain.Presence{}, &err
	}

	presence, ok, err := s.store.GetPresence(playerID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Presence{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "presence not found", http.StatusNotFound)
		return domain.Presence{}, &err
	}

	return presence, nil
}

func (s *PresenceService) String() string {
	return fmt.Sprintf("presence-service(store=%T)", s.store)
}
