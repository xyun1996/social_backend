package jobs

import (
	"context"
	"encoding/json"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

// OfflineDeliveryRecorder handles chat-side offline delivery processing.
type OfflineDeliveryRecorder interface {
	RecordOfflineDelivery(ctx context.Context, payload map[string]any) *apperrors.Error
}

// ChatOfflineDeliveryHandler executes offline chat delivery jobs.
type ChatOfflineDeliveryHandler struct {
	chat OfflineDeliveryRecorder
}

// NewChatOfflineDeliveryHandler constructs the chat offline delivery handler.
func NewChatOfflineDeliveryHandler(chat OfflineDeliveryRecorder) *ChatOfflineDeliveryHandler {
	return &ChatOfflineDeliveryHandler{chat: chat}
}

// Handle runs the chat offline delivery job.
func (h *ChatOfflineDeliveryHandler) Handle(ctx context.Context, job domain.Job) *apperrors.Error {
	if h.chat == nil {
		err := apperrors.New("dependency_missing", "chat recorder is not configured", 500)
		return &err
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		appErr := apperrors.New("invalid_payload", "chat.offline_delivery payload must be valid json", 400)
		return &appErr
	}

	return h.chat.RecordOfflineDelivery(ctx, payload)
}
