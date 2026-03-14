package jobs

import (
	"context"
	"encoding/json"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

// GuildActivityPeriodMaintainer handles guild progression period maintenance.
type GuildActivityPeriodMaintainer interface {
	EnsureCurrentActivityInstances(ctx context.Context, guildID string) *apperrors.Error
	CloseExpiredActivityInstances(ctx context.Context, guildID string) *apperrors.Error
}

type guildActivityPayload struct {
	GuildID string `json:"guild_id"`
}

// GuildActivityEnsureHandler executes guild instance initialization jobs.
type GuildActivityEnsureHandler struct {
	guilds GuildActivityPeriodMaintainer
}

// NewGuildActivityEnsureHandler constructs the ensure-current handler.
func NewGuildActivityEnsureHandler(guilds GuildActivityPeriodMaintainer) *GuildActivityEnsureHandler {
	return &GuildActivityEnsureHandler{guilds: guilds}
}

// Handle runs the ensure-current guild job.
func (h *GuildActivityEnsureHandler) Handle(ctx context.Context, job domain.Job) *apperrors.Error {
	payload, appErr := decodeGuildActivityPayload(job.Payload)
	if appErr != nil {
		return appErr
	}
	if h.guilds == nil {
		err := apperrors.New("dependency_missing", "guild maintainer is not configured", 500)
		return &err
	}
	return h.guilds.EnsureCurrentActivityInstances(ctx, payload.GuildID)
}

// GuildActivityCloseHandler executes guild expired-instance transitions.
type GuildActivityCloseHandler struct {
	guilds GuildActivityPeriodMaintainer
}

// NewGuildActivityCloseHandler constructs the close-expired handler.
func NewGuildActivityCloseHandler(guilds GuildActivityPeriodMaintainer) *GuildActivityCloseHandler {
	return &GuildActivityCloseHandler{guilds: guilds}
}

// Handle runs the close-expired guild job.
func (h *GuildActivityCloseHandler) Handle(ctx context.Context, job domain.Job) *apperrors.Error {
	payload, appErr := decodeGuildActivityPayload(job.Payload)
	if appErr != nil {
		return appErr
	}
	if h.guilds == nil {
		err := apperrors.New("dependency_missing", "guild maintainer is not configured", 500)
		return &err
	}
	return h.guilds.CloseExpiredActivityInstances(ctx, payload.GuildID)
}

func decodeGuildActivityPayload(raw string) (guildActivityPayload, *apperrors.Error) {
	var payload guildActivityPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		appErr := apperrors.New("invalid_payload", "guild activity payload must be valid json", 400)
		return guildActivityPayload{}, &appErr
	}
	if payload.GuildID == "" {
		appErr := apperrors.New("invalid_payload", "guild activity payload requires guild_id", 400)
		return guildActivityPayload{}, &appErr
	}
	return payload, nil
}
