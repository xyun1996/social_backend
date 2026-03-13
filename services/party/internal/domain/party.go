package domain

import "time"

// Party is the root aggregate for the in-memory party prototype.
type Party struct {
	ID        string    `json:"id"`
	LeaderID  string    `json:"leader_id"`
	MemberIDs []string  `json:"member_ids"`
	CreatedAt time.Time `json:"created_at"`
}

// ReadyState tracks a member's readiness inside a party.
type ReadyState struct {
	PartyID   string    `json:"party_id"`
	PlayerID  string    `json:"player_id"`
	IsReady   bool      `json:"is_ready"`
	UpdatedAt time.Time `json:"updated_at"`
}
