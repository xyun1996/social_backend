package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

type fakePresenceReader struct{}

func (f *fakePresenceReader) GetPresence(context.Context, string) (opsservice.PresenceRecord, *apperrors.Error) {
	return opsservice.PresenceRecord{PlayerID: "p1", Status: "online"}, nil
}

type fakePartyReader struct{}

func (f *fakePartyReader) GetPartySnapshot(context.Context, string) (opsservice.PartySnapshot, *apperrors.Error) {
	return opsservice.PartySnapshot{PartyID: "party-1", Count: 1}, nil
}

type fakeGuildReader struct{}

func (f *fakeGuildReader) GetGuildSnapshot(context.Context, string) (opsservice.GuildSnapshot, *apperrors.Error) {
	return opsservice.GuildSnapshot{GuildID: "guild-1", Count: 1}, nil
}

func TestOpsEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(opsservice.NewOpsService(&fakePresenceReader{}, &fakePartyReader{}, &fakeGuildReader{}))

	tests := []string{
		"/v1/ops/players/p1/presence",
		"/v1/ops/parties/party-1",
		"/v1/ops/guilds/guild-1",
	}

	for _, path := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status for %s: got %d want %d", path, rec.Code, http.StatusOK)
		}
	}
}
