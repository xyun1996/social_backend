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
	return opsservice.PartySnapshot{
		PartyID: "party-1",
		Count:   1,
		Queue: &opsservice.PartyQueueState{
			PartyID:   "party-1",
			QueueName: "casual-2v2",
			Status:    "queued",
		},
	}, nil
}

type fakeGuildReader struct{}

func (f *fakeGuildReader) GetGuildSnapshot(context.Context, string) (opsservice.GuildSnapshot, *apperrors.Error) {
	return opsservice.GuildSnapshot{GuildID: "guild-1", Count: 1}, nil
}

type fakeWorkerReader struct{}

func (f *fakeWorkerReader) GetWorkerSnapshot(context.Context, string, string) (opsservice.WorkerSnapshot, *apperrors.Error) {
	return opsservice.WorkerSnapshot{Count: 1}, nil
}

type fakeSocialReader struct{}

func (f *fakeSocialReader) GetSocialSnapshot(context.Context, string) (opsservice.SocialSnapshot, *apperrors.Error) {
	return opsservice.SocialSnapshot{PlayerID: "p1", Friends: []string{"p2"}, Blocks: []string{"p3"}}, nil
}

type fakeBootstrapReader struct{}

func (f *fakeBootstrapReader) GetMySQLBootstrapSnapshot(context.Context) (opsservice.MySQLBootstrapSnapshot, *apperrors.Error) {
	return opsservice.MySQLBootstrapSnapshot{
		Count: 1,
		Services: []opsservice.MySQLBootstrapService{
			{Service: "chat", Count: 1, MigrationIDs: []string{"001_chat_core"}},
		},
	}, nil
}

type fakeRedisRuntimeReader struct{}

func (f *fakeRedisRuntimeReader) GetRedisRuntimeSnapshot(context.Context) (opsservice.RedisRuntimeSnapshot, *apperrors.Error) {
	return opsservice.RedisRuntimeSnapshot{
		PresenceRecordCount: 1,
		GatewaySessionCount: 1,
		WorkerJobCount:      1,
	}, nil
}

func TestOpsEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(opsservice.NewOpsService(&fakePresenceReader{}, &fakePartyReader{}, &fakeGuildReader{}, &fakeWorkerReader{}, &fakeSocialReader{}, &fakeBootstrapReader{}, &fakeRedisRuntimeReader{}))

	tests := []string{
		"/v1/ops/players/p1/overview",
		"/v1/ops/players/p1/presence",
		"/v1/ops/parties/party-1",
		"/v1/ops/guilds/guild-1",
		"/v1/ops/jobs?status=queued",
		"/v1/ops/durable/summary",
		"/v1/ops/bootstrap/mysql",
		"/v1/ops/runtime/redis",
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
