package domain

import "time"

// Conversation represents a chat container with explicit member scope.
type Conversation struct {
	ID              string    `json:"id"`
	Kind            string    `json:"kind"`
	ResourceID      string    `json:"resource_id,omitempty"`
	MemberPlayerIDs []string  `json:"member_player_ids"`
	LastSeq         int64     `json:"last_seq"`
	CreatedAt       time.Time `json:"created_at"`
}

// ChannelDescriptor explains the policy surface implied by a conversation kind.
type ChannelDescriptor struct {
	ConversationID   string `json:"conversation_id"`
	Kind             string `json:"kind"`
	ResourceID       string `json:"resource_id,omitempty"`
	Scope            string `json:"scope"`
	MembershipMode   string `json:"membership_mode"`
	SendPolicy       string `json:"send_policy"`
	ResourceRequired bool   `json:"resource_required"`
	MemberCount      int    `json:"member_count"`
}

// Message is an immutable conversation entry ordered by seq.
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Seq            int64     `json:"seq"`
	SenderPlayerID string    `json:"sender_player_id"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

// ReadCursor tracks the highest acknowledged seq for a player in a conversation.
type ReadCursor struct {
	ConversationID string    `json:"conversation_id"`
	PlayerID       string    `json:"player_id"`
	AckSeq         int64     `json:"ack_seq"`
	UpdatedAt      time.Time `json:"updated_at"`
}
