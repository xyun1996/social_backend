package guild

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGuildSnapshotIncludesAnnouncementAndLogs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/guilds/guild-1":
			_, _ = w.Write([]byte(`{"id":"guild-1","name":"Raiders","owner_id":"p1","announcement":"Welcome","announcement_updated_at":"2026-03-13T10:05:00Z"}`))
		case "/v1/guilds/guild-1/members":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","count":2,"members":[{"player_id":"p1","role":"owner","presence":"online"},{"player_id":"p2","role":"member","presence":"offline"}]}`))
		case "/v1/guilds/guild-1/logs":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","count":1,"logs":[{"id":"log-1","action":"guild.created","actor_id":"p1","message":"guild created","created_at":"2026-03-13T10:00:00Z"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	record, err := client.GetGuildSnapshot(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("GetGuildSnapshot returned error: %+v", err)
	}
	if record.Name != "Raiders" || record.Announcement != "Welcome" {
		t.Fatalf("unexpected guild aggregate fields: %+v", record)
	}
	if record.LogCount != 1 || len(record.Logs) != 1 || record.Logs[0].Action != "guild.created" {
		t.Fatalf("unexpected guild logs: %+v", record)
	}
}
