package domain

import "time"

// Presence tracks online state and minimal runtime metadata for a player.
type Presence struct {
	PlayerID        string     `json:"player_id"`
	Status          string     `json:"status"`
	SessionID       string     `json:"session_id"`
	RealmID         string     `json:"realm_id,omitempty"`
	Location        string     `json:"location,omitempty"`
	LastHeartbeatAt time.Time  `json:"last_heartbeat_at"`
	LastSeenAt      time.Time  `json:"last_seen_at"`
	ConnectedAt     *time.Time `json:"connected_at,omitempty"`
	DisconnectedAt  *time.Time `json:"disconnected_at,omitempty"`
}
