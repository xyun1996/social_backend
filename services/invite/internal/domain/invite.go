package domain

import "time"

// Invite captures the first cross-domain invitation lifecycle model.
type Invite struct {
	ID           string     `json:"id"`
	Domain       string     `json:"domain"`
	ResourceID   string     `json:"resource_id,omitempty"`
	FromPlayerID string     `json:"from_player_id"`
	ToPlayerID   string     `json:"to_player_id"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	RespondedAt  *time.Time `json:"responded_at,omitempty"`
}
