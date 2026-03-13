package domain

import "time"

// Guild is the root aggregate for the in-memory guild prototype.
type Guild struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	OwnerID   string        `json:"owner_id"`
	Members   []GuildMember `json:"members"`
	CreatedAt time.Time     `json:"created_at"`
}

// GuildMember captures a member's role in the guild.
type GuildMember struct {
	PlayerID string    `json:"player_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
