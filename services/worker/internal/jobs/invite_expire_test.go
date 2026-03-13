package jobs

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
)

type fakeInviteExpirer struct {
	inviteID string
	err      *apperrors.Error
}

func (f *fakeInviteExpirer) ExpireInvite(_ context.Context, inviteID string) *apperrors.Error {
	f.inviteID = inviteID
	return f.err
}

func TestInviteExpireHandler(t *testing.T) {
	t.Parallel()

	expirer := &fakeInviteExpirer{}
	handler := NewInviteExpireHandler(expirer)

	appErr := handler.Handle(context.Background(), domain.Job{
		Type:    "invite.expire",
		Payload: `{"invite_id":"inv-1"}`,
	})
	if appErr != nil {
		t.Fatalf("handle returned error: %+v", appErr)
	}
	if expirer.inviteID != "inv-1" {
		t.Fatalf("unexpected invite id: %q", expirer.inviteID)
	}
}
