package domain

import "time"

// FriendRequest tracks a directed friend application before acceptance.
type FriendRequest struct {
	ID           string    `json:"id"`
	FromPlayerID string    `json:"from_player_id"`
	ToPlayerID   string    `json:"to_player_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// FriendRelationship represents a confirmed bidirectional friendship.
type FriendRelationship struct {
	PlayerID string `json:"player_id"`
	FriendID string `json:"friend_id"`
}

// FriendRemark stores optional player-defined metadata for a confirmed friend.
type FriendRemark struct {
	PlayerID   string    `json:"player_id"`
	FriendID   string    `json:"friend_id"`
	Remark     string    `json:"remark"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RelationshipSnapshot is the richer point-to-point social read surface added in v2.
type RelationshipSnapshot struct {
	PlayerID         string    `json:"player_id"`
	TargetPlayerID   string    `json:"target_player_id"`
	State            string    `json:"state"`
	IsFriend         bool      `json:"is_friend"`
	HasPendingInbox  bool      `json:"has_pending_inbox"`
	HasPendingOutbox bool      `json:"has_pending_outbox"`
	IsBlocked        bool      `json:"is_blocked"`
	IsBlockedBy      bool      `json:"is_blocked_by"`
	Remark           string    `json:"remark,omitempty"`
	ReverseRemark    string    `json:"reverse_remark,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// PendingSummary aggregates inbox and outbox pending request state for one player.
type PendingSummary struct {
	PlayerID      string   `json:"player_id"`
	Inbox         []string `json:"inbox"`
	Outbox        []string `json:"outbox"`
	InboxCount    int      `json:"inbox_count"`
	OutboxCount   int      `json:"outbox_count"`
	TotalPending  int      `json:"total_pending"`
}
