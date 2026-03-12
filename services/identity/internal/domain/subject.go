package domain

// Subject describes the authenticated player context resolved from an access token.
type Subject struct {
	AccountID string `json:"account_id"`
	PlayerID  string `json:"player_id"`
}
