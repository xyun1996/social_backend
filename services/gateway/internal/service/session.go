package service

import (
	"context"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// Subject is the authenticated player context returned by the identity boundary.
type Subject struct {
	AccountID string `json:"account_id"`
	PlayerID  string `json:"player_id"`
}

// Introspector resolves bearer tokens into authenticated subjects.
type Introspector interface {
	Introspect(ctx context.Context, accessToken string) (Subject, *apperrors.Error)
}
