package jobs

import (
	"context"
	"encoding/json"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

// InviteExpirer handles the invite expiry side effect.
type InviteExpirer interface {
	ExpireInvite(ctx context.Context, inviteID string) *apperrors.Error
}

// InviteExpireHandler executes invite expiry jobs.
type InviteExpireHandler struct {
	invites InviteExpirer
}

// NewInviteExpireHandler constructs the invite expiry job handler.
func NewInviteExpireHandler(invites InviteExpirer) *InviteExpireHandler {
	return &InviteExpireHandler{invites: invites}
}

type inviteExpirePayload struct {
	InviteID string `json:"invite_id"`
}

// Handle runs the invite expiry job.
func (h *InviteExpireHandler) Handle(ctx context.Context, job domain.Job) *apperrors.Error {
	if h.invites == nil {
		err := apperrors.New("dependency_missing", "invite expirer is not configured", 500)
		return &err
	}

	var payload inviteExpirePayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		appErr := apperrors.New("invalid_payload", "invite.expire payload must be valid json", 400)
		return &appErr
	}
	if payload.InviteID == "" {
		appErr := apperrors.New("invalid_payload", "invite.expire payload requires invite_id", 400)
		return &appErr
	}

	return h.invites.ExpireInvite(ctx, payload.InviteID)
}
