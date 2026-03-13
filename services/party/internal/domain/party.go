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

// QueueState tracks the current social queue enrollment for a party.
type QueueState struct {
	PartyID   string    `json:"party_id"`
	QueueName string    `json:"queue_name"`
	Status    string    `json:"status"`
	JoinedBy  string    `json:"joined_by"`
	JoinedAt  time.Time `json:"joined_at"`
}

// QueueLeaveResult describes a successful queue exit.
type QueueLeaveResult struct {
	PartyID   string    `json:"party_id"`
	QueueName string    `json:"queue_name"`
	Status    string    `json:"status"`
	LeftAt    time.Time `json:"left_at"`
}

// QueueHandoff is the stable queue payload exposed to an external matchmaker boundary.
type QueueHandoff struct {
	TicketID  string    `json:"ticket_id"`
	PartyID   string    `json:"party_id"`
	QueueName string    `json:"queue_name"`
	LeaderID  string    `json:"leader_id"`
	MemberIDs []string  `json:"member_ids"`
	JoinedAt  time.Time `json:"joined_at"`
}
