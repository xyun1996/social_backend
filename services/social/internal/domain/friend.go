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
