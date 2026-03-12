package domain

import "time"

// TokenPair represents the issued access and refresh tokens for a player session.
type TokenPair struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    time.Duration `json:"expires_in_seconds"`
	AccountID    string        `json:"account_id"`
	PlayerID     string        `json:"player_id"`
}

// Session tracks the minimal identity session state needed by the in-memory prototype.
type Session struct {
	AccountID    string
	PlayerID     string
	RefreshToken string
	ExpiresAt    time.Time
}
