package service

import (
	"context"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// Subject is the authenticated player context returned by the identity boundary.
type Subject struct {
	AccountID string `json:"account_id"`
	PlayerID  string `json:"player_id"`
}

// Introspector resolves bearer tokens into authenticated subjects.
type Introspector interface {
	Introspect(ctx context.Context, accessToken string) (Subject, *apperrors.Error)
}

// PresenceUpdate is the gateway-owned runtime metadata forwarded to presence.
type PresenceUpdate struct {
	PlayerID  string `json:"player_id"`
	SessionID string `json:"session_id"`
	RealmID   string `json:"realm_id,omitempty"`
	Location  string `json:"location,omitempty"`
}

// PresenceSnapshot is the presence state returned to gateway callers.
type PresenceSnapshot struct {
	PlayerID        string `json:"player_id"`
	Status          string `json:"status"`
	SessionID       string `json:"session_id"`
	RealmID         string `json:"realm_id,omitempty"`
	Location        string `json:"location,omitempty"`
	LastHeartbeatAt string `json:"last_heartbeat_at"`
	LastSeenAt      string `json:"last_seen_at"`
	ConnectedAt     string `json:"connected_at,omitempty"`
	DisconnectedAt  string `json:"disconnected_at,omitempty"`
}

// PresenceReporter forwards gateway lifecycle updates to the presence boundary.
type PresenceReporter interface {
	Connect(ctx context.Context, update PresenceUpdate) (PresenceSnapshot, *apperrors.Error)
	Heartbeat(ctx context.Context, update PresenceUpdate) (PresenceSnapshot, *apperrors.Error)
	Disconnect(ctx context.Context, update PresenceUpdate) (PresenceSnapshot, *apperrors.Error)
}
