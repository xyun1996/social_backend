package domain

import "time"

// BlockRelationship suppresses point-to-point interactions from blocker to blocked player.
type BlockRelationship struct {
	PlayerID  string    `json:"player_id"`
	BlockedID string    `json:"blocked_id"`
	CreatedAt time.Time `json:"created_at"`
}
