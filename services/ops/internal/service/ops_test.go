package service

import (
	"context"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

type fakePresenceReader struct {
	record PresenceRecord
	err    *apperrors.Error
}

func (f *fakePresenceReader) GetPresence(context.Context, string) (PresenceRecord, *apperrors.Error) {
	return f.record, f.err
}

type fakePartyReader struct {
	record PartySnapshot
	err    *apperrors.Error
}

func (f *fakePartyReader) GetPartySnapshot(context.Context, string) (PartySnapshot, *apperrors.Error) {
	return f.record, f.err
}

type fakeGuildReader struct {
	record GuildSnapshot
	err    *apperrors.Error
}

func (f *fakeGuildReader) GetGuildSnapshot(context.Context, string) (GuildSnapshot, *apperrors.Error) {
	return f.record, f.err
}

type fakeWorkerReader struct {
	record WorkerSnapshot
	err    *apperrors.Error
}

func (f *fakeWorkerReader) GetWorkerSnapshot(context.Context, string, string) (WorkerSnapshot, *apperrors.Error) {
	return f.record, f.err
}

type fakeSocialReader struct {
	record SocialSnapshot
	err    *apperrors.Error
}

func (f *fakeSocialReader) GetSocialSnapshot(context.Context, string) (SocialSnapshot, *apperrors.Error) {
	return f.record, f.err
}

type fakeBootstrapReader struct {
	record MySQLBootstrapSnapshot
	err    *apperrors.Error
}

func (f *fakeBootstrapReader) GetMySQLBootstrapSnapshot(context.Context) (MySQLBootstrapSnapshot, *apperrors.Error) {
	return f.record, f.err
}

type fakeRedisRuntimeReader struct {
	record RedisRuntimeSnapshot
	err    *apperrors.Error
}

func (f *fakeRedisRuntimeReader) GetRedisRuntimeSnapshot(context.Context) (RedisRuntimeSnapshot, *apperrors.Error) {
	return f.record, f.err
}

func TestGetPlayerPresence(t *testing.T) {
	t.Parallel()

	svc := NewOpsService(&fakePresenceReader{
		record: PresenceRecord{PlayerID: "p1", Status: "online"},
	}, nil, nil, nil, nil, nil, nil)

	record, err := svc.GetPlayerPresence(context.Background(), "p1")
	if err != nil {
		t.Fatalf("get presence returned error: %+v", err)
	}
	if record.PlayerID != "p1" || record.Status != "online" {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestGetPartySnapshot(t *testing.T) {
	t.Parallel()

	svc := NewOpsService(nil, &fakePartyReader{
		record: PartySnapshot{PartyID: "party-1", Count: 1},
	}, nil, nil, nil, nil, nil)

	record, err := svc.GetPartySnapshot(context.Background(), "party-1")
	if err != nil {
		t.Fatalf("get party returned error: %+v", err)
	}
	if record.PartyID != "party-1" {
		t.Fatalf("unexpected party snapshot: %+v", record)
	}
}

func TestGetWorkerSnapshot(t *testing.T) {
	t.Parallel()

	svc := NewOpsService(nil, nil, nil, &fakeWorkerReader{
		record: WorkerSnapshot{Count: 1},
	}, nil, nil, nil)

	record, err := svc.GetWorkerSnapshot(context.Background(), "queued", "invite.expire")
	if err != nil {
		t.Fatalf("get worker returned error: %+v", err)
	}
	if record.Count != 1 {
		t.Fatalf("unexpected worker snapshot: %+v", record)
	}
}

func TestGetPlayerOverview(t *testing.T) {
	t.Parallel()

	svc := NewOpsService(
		&fakePresenceReader{record: PresenceRecord{PlayerID: "p1", Status: "online"}},
		nil,
		nil,
		nil,
		&fakeSocialReader{record: SocialSnapshot{PlayerID: "p1", Friends: []string{"p2"}, Blocks: []string{"p3"}, PendingInbox: []string{"p4"}, PendingOutbox: []string{"p5"}}},
		nil,
		nil,
	)

	record, err := svc.GetPlayerOverview(context.Background(), "p1")
	if err != nil {
		t.Fatalf("get player overview returned error: %+v", err)
	}
	if record.PlayerID != "p1" || record.FriendCnt != 1 || record.BlockCnt != 1 || record.PendingInboxCount != 1 || record.PendingOutboxCount != 1 {
		t.Fatalf("unexpected player overview: %+v", record)
	}
}

func TestGetMySQLBootstrapSnapshot(t *testing.T) {
	t.Parallel()

	svc := NewOpsService(nil, nil, nil, nil, nil, &fakeBootstrapReader{
		record: MySQLBootstrapSnapshot{
			Count: 1,
			Services: []MySQLBootstrapService{
				{Service: "chat", Count: 1, MigrationIDs: []string{"001_chat_core"}},
			},
		},
	}, nil)

	record, err := svc.GetMySQLBootstrapSnapshot(context.Background())
	if err != nil {
		t.Fatalf("get mysql bootstrap returned error: %+v", err)
	}
	if record.Count != 1 || len(record.Services) != 1 || record.Services[0].Service != "chat" {
		t.Fatalf("unexpected mysql bootstrap snapshot: %+v", record)
	}
}

func TestGetRedisRuntimeSnapshot(t *testing.T) {
	t.Parallel()

	svc := NewOpsService(nil, nil, nil, nil, nil, nil, &fakeRedisRuntimeReader{
		record: RedisRuntimeSnapshot{
			PresenceRecordCount: 1,
			GatewaySessionCount: 1,
			WorkerJobCount:      2,
			WorkerStatusCounters: []RedisWorkerStatusCount{
				{Status: "queued", Count: 2},
			},
		},
	})

	record, err := svc.GetRedisRuntimeSnapshot(context.Background())
	if err != nil {
		t.Fatalf("get redis runtime returned error: %+v", err)
	}
	if record.PresenceRecordCount != 1 || record.GatewaySessionCount != 1 || record.WorkerJobCount != 2 {
		t.Fatalf("unexpected redis runtime snapshot: %+v", record)
	}
}
