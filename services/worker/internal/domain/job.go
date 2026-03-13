package domain

import "time"

// Job is the in-memory worker unit for async execution and compensation flows.
type Job struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Payload     string     `json:"payload"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempts"`
	LastError   string     `json:"last_error,omitempty"`
	ClaimedBy   string     `json:"claimed_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
