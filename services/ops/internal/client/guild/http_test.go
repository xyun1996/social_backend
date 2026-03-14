package guild

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGuildSnapshotIncludesProgressionAndRewards(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/guilds/guild-1":
			_, _ = w.Write([]byte(`{"id":"guild-1","name":"Raiders","owner_id":"p1","announcement":"Welcome","announcement_updated_at":"2026-03-13T10:05:00Z"}`))
		case "/v1/guilds/guild-1/members":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","count":2,"members":[{"player_id":"p1","role":"owner","presence":"online"},{"player_id":"p2","role":"member","presence":"offline"}]}`))
		case "/v1/guilds/guild-1/logs":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","count":1,"logs":[{"id":"log-1","action":"guild.created","actor_id":"p1","message":"guild created","created_at":"2026-03-13T10:00:00Z"}]}`))
		case "/v1/guilds/guild-1/progression":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","level":2,"experience":125,"next_level_xp":200}`))
		case "/v1/guilds/guild-1/contributions":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","count":1,"contributions":[{"player_id":"p1","total_xp":125,"last_source_type":"donate","updated_at":"2026-03-13T10:10:00Z"}]}`))
		case "/v1/guilds/guild-1/rewards":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","count":1,"rewards":[{"id":"reward-1","player_id":"p1","activity_id":"act-1","template_key":"donate","reward_type":"token","reward_ref":"guild_donation","created_at":"2026-03-13T10:10:00Z"}]}`))
		case "/v1/guilds/guild-1/activities/sign_in/instances":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","template_key":"sign_in","count":1,"instances":[{"id":"inst-1","template_key":"sign_in","period_key":"2026-03-13","status":"active","starts_at":"2026-03-13T00:00:00Z","ends_at":"2026-03-14T00:00:00Z","updated_at":"2026-03-13T00:00:00Z"}]}`))
		case "/v1/guilds/guild-1/activities/donate/instances", "/v1/guilds/guild-1/activities/guild_task/instances":
			_, _ = w.Write([]byte(`{"guild_id":"guild-1","count":0,"instances":[]}`))
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
	if record.Name != "Raiders" || record.Announcement != "Welcome" || record.Level != 2 || record.Experience != 125 {
		t.Fatalf("unexpected guild aggregate fields: %+v", record)
	}
	if record.LogCount != 1 || len(record.Logs) != 1 || record.Logs[0].Action != "guild.created" {
		t.Fatalf("unexpected guild logs: %+v", record)
	}
	if len(record.Contributions) != 1 || record.Contributions[0].TotalXP != 125 {
		t.Fatalf("unexpected guild contributions: %+v", record)
	}
	if len(record.ActivityInstances) != 1 || record.ActivityInstances[0].TemplateKey != "sign_in" {
		t.Fatalf("unexpected guild instances: %+v", record)
	}
	if len(record.RewardRecords) != 1 || record.RewardRecords[0].RewardType != "token" {
		t.Fatalf("unexpected guild rewards: %+v", record)
	}
}
