package domain

import "time"

// Guild is the root aggregate for the in-memory guild prototype.
type Guild struct {
	ID                    string        `json:"id"`
	Name                  string        `json:"name"`
	OwnerID               string        `json:"owner_id"`
	Level                 int           `json:"level"`
	Experience            int           `json:"experience"`
	Announcement          string        `json:"announcement,omitempty"`
	AnnouncementUpdatedAt time.Time     `json:"announcement_updated_at,omitempty"`
	Members               []GuildMember `json:"members"`
	CreatedAt             time.Time     `json:"created_at"`
}

// GuildMember captures a member's role in the guild.
type GuildMember struct {
	PlayerID string    `json:"player_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// GuildLogEntry captures a governance event in guild history.
type GuildLogEntry struct {
	ID        string    `json:"id"`
	GuildID   string    `json:"guild_id"`
	Action    string    `json:"action"`
	ActorID   string    `json:"actor_id,omitempty"`
	TargetID  string    `json:"target_id,omitempty"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// GuildActivityTemplate is a fixed activity definition shipped with the guild domain.
type GuildActivityTemplate struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	ContributionXP int    `json:"contribution_xp"`
}

// GuildActivityRecord is a submitted activity action for a guild member.
type GuildActivityRecord struct {
	ID          string    `json:"id"`
	GuildID     string    `json:"guild_id"`
	TemplateKey string    `json:"template_key"`
	PlayerID    string    `json:"player_id"`
	DeltaXP     int       `json:"delta_xp"`
	CreatedAt   time.Time `json:"created_at"`
}
