package jobs

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

type fakeOfflineDeliveryRecorder struct {
	payload map[string]any
	err     *apperrors.Error
}

func (f *fakeOfflineDeliveryRecorder) RecordOfflineDelivery(_ context.Context, payload map[string]any) *apperrors.Error {
	f.payload = payload
	return f.err
}

func TestChatOfflineDeliveryHandler(t *testing.T) {
	t.Parallel()

	recorder := &fakeOfflineDeliveryRecorder{}
	handler := NewChatOfflineDeliveryHandler(recorder)

	appErr := handler.Handle(context.Background(), domain.Job{
		Type:    "chat.offline_delivery",
		Payload: `{"conversation_id":"conv-1","message_id":"msg-1","recipient_player":"p2"}`,
	})
	if appErr != nil {
		t.Fatalf("handle returned error: %+v", appErr)
	}
	if recorder.payload["message_id"] != "msg-1" {
		t.Fatalf("unexpected payload: %+v", recorder.payload)
	}
}
