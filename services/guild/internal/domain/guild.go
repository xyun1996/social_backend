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
	Key                     string `json:"key"`
	Name                    string `json:"name"`
	PeriodType              string `json:"period_type"`
	MaxSubmissionsPerPeriod int    `json:"max_submissions_per_period"`
	ContributionXP          int    `json:"contribution_xp"`
	RewardType              string `json:"reward_type,omitempty"`
	RewardRef               string `json:"reward_ref,omitempty"`
}

// GuildProgression is the read model for guild level growth.
type GuildProgression struct {
	GuildID     string    `json:"guild_id"`
	Level       int       `json:"level"`
	Experience  int       `json:"experience"`
	NextLevelXP int       `json:"next_level_xp"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GuildContribution tracks one member's cumulative guild contribution.
type GuildContribution struct {
	GuildID        string    `json:"guild_id"`
	PlayerID       string    `json:"player_id"`
	TotalXP        int       `json:"total_xp"`
	LastSourceType string    `json:"last_source_type,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GuildActivityInstance is the active or historical period container for a template.
type GuildActivityInstance struct {
	ID          string    `json:"id"`
	GuildID     string    `json:"guild_id"`
	TemplateKey string    `json:"template_key"`
	PeriodKey   string    `json:"period_key"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GuildActivityRecord is a submitted activity action for a guild member.
type GuildActivityRecord struct {
	ID             string    `json:"id"`
	InstanceID     string    `json:"instance_id"`
	GuildID        string    `json:"guild_id"`
	TemplateKey    string    `json:"template_key"`
	PlayerID       string    `json:"player_id"`
	DeltaXP        int       `json:"delta_xp"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	SourceType     string    `json:"source_type,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// GuildRewardRecord stores reward bookkeeping without external fulfillment orchestration.
type GuildRewardRecord struct {
	ID          string    `json:"id"`
	GuildID     string    `json:"guild_id"`
	PlayerID    string    `json:"player_id"`
	ActivityID  string    `json:"activity_id"`
	TemplateKey string    `json:"template_key"`
	RewardType  string    `json:"reward_type"`
	RewardRef   string    `json:"reward_ref,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
