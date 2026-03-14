package integration

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	chattestkit "github.com/xyun1996/social_backend/services/chat/testkit"
	gatewaytestkit "github.com/xyun1996/social_backend/services/gateway/testkit"
	guildtestkit "github.com/xyun1996/social_backend/services/guild/testkit"
	identitytestkit "github.com/xyun1996/social_backend/services/identity/testkit"
	invitetestkit "github.com/xyun1996/social_backend/services/invite/testkit"
	opstestkit "github.com/xyun1996/social_backend/services/ops/testkit"
	partytestkit "github.com/xyun1996/social_backend/services/party/testkit"
	presencetestkit "github.com/xyun1996/social_backend/services/presence/testkit"
	socialtestkit "github.com/xyun1996/social_backend/services/social/testkit"
	workertestkit "github.com/xyun1996/social_backend/services/worker/testkit"

	_ "github.com/go-sql-driver/mysql"
)

func TestLocalDurableChatAndInviteFlows(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	chat := chattestkit.NewDurableServer(mysqlConfig, sqlDB, "", "")
	defer chat.Close()
	invite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, "")
	defer invite.Close()

	var conversation struct {
		ID string `json:"id"`
	}
	postJSON(t, chat.URL()+"/v1/conversations", map[string]any{
		"kind":              "private",
		"member_player_ids": []string{"p1", "p2"},
	}, &conversation)

	var message struct {
		Seq  int64  `json:"seq"`
		Body string `json:"body"`
	}
	postJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/messages", map[string]any{
		"sender_player_id": "p1",
		"body":             "hello durable",
	}, &message)
	if message.Seq != 1 {
		t.Fatalf("unexpected durable chat message: %+v", message)
	}

	postJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/ack", map[string]any{
		"player_id": "p2",
		"ack_seq":   1,
	}, nil)

	var replay struct {
		Count    int `json:"count"`
		Messages []struct {
			Body string `json:"body"`
		} `json:"messages"`
	}
	getJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/messages?player_id=p2&after_seq=0", &replay)
	if replay.Count != 1 || replay.Messages[0].Body != "hello durable" {
		t.Fatalf("unexpected durable replay payload: %+v", replay)
	}

	var createdInvite struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	postJSON(t, invite.URL()+"/v1/invites", map[string]any{
		"domain":         "party",
		"resource_id":    "party-1",
		"from_player_id": "p1",
		"to_player_id":   "p2",
		"ttl_seconds":    300,
	}, &createdInvite)

	var acceptedInvite struct {
		Status string `json:"status"`
	}
	postJSON(t, invite.URL()+"/v1/invites/"+createdInvite.ID+"/accept", map[string]any{
		"actor_player_id": "p2",
	}, &acceptedInvite)
	if acceptedInvite.Status != "accepted" {
		t.Fatalf("unexpected durable invite state: %+v", acceptedInvite)
	}
}

func TestLocalDurableSocialPresenceWorkerFlows(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	social := socialtestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer social.Close()

	redisConfig, redisClient := newLocalRedisTestClient(t)
	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()
	worker := workertestkit.NewDurableServer(redisConfig, redisClient)
	defer worker.Close()

	var request struct {
		ID string `json:"id"`
	}
	postJSON(t, social.URL()+"/v1/friends/requests", map[string]any{
		"from_player_id": "p1",
		"to_player_id":   "p2",
	}, &request)
	postJSON(t, social.URL()+"/v1/friends/requests/"+request.ID+"/accept", map[string]any{
		"actor_player_id": "p2",
	}, nil)

	var friends struct {
		Friends []string `json:"friends"`
	}
	getJSON(t, social.URL()+"/v1/friends?player_id=p1", &friends)
	if len(friends.Friends) != 1 || friends.Friends[0] != "p2" {
		t.Fatalf("unexpected durable friends payload: %+v", friends)
	}

	postJSON(t, presence.URL()+"/v1/presence/connect", map[string]any{
		"player_id":  "p1",
		"session_id": "sess-1",
		"realm_id":   "realm-1",
		"location":   "lobby",
	}, nil)

	var presencePayload struct {
		Status string `json:"status"`
	}
	getJSON(t, presence.URL()+"/v1/presence/p1", &presencePayload)
	if presencePayload.Status != "online" {
		t.Fatalf("unexpected durable presence payload: %+v", presencePayload)
	}

	postJSON(t, worker.URL()+"/v1/jobs", map[string]any{
		"type":    "invite.expire",
		"payload": `{"invite_id":"inv-1"}`,
	}, nil)

	var jobs struct {
		Count int `json:"count"`
	}
	getJSON(t, worker.URL()+"/v1/jobs", &jobs)
	if jobs.Count != 1 {
		t.Fatalf("unexpected durable job payload: %+v", jobs)
	}
}

func TestLocalDurableGatewayHandshakeFlow(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	identity := identitytestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer identity.Close()

	redisConfig, redisClient := newLocalRedisTestClient(t)
	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()

	gateway := gatewaytestkit.NewServer(identity.URL(), presence.URL(), "")
	defer gateway.Close()

	var tokenPair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
		"account_id": "a1",
		"player_id":  "p1",
	}, &tokenPair)

	postJSON(t, gateway.URL()+"/v1/realtime/handshake", map[string]any{
		"access_token": tokenPair.AccessToken,
		"session_id":   "sess-1",
		"realm_id":     "realm-1",
		"location":     "lobby",
	}, nil)

	var presencePayload struct {
		PlayerID  string `json:"player_id"`
		Status    string `json:"status"`
		SessionID string `json:"session_id"`
	}
	getJSON(t, presence.URL()+"/v1/presence/p1", &presencePayload)
	if presencePayload.PlayerID != "p1" || presencePayload.Status != "online" || presencePayload.SessionID != "sess-1" {
		t.Fatalf("unexpected durable handshake presence payload: %+v", presencePayload)
	}
}

func TestLocalDurableInviteWorkerExpiryFlow(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	redisConfig, redisClient := newLocalRedisTestClient(t)

	worker := workertestkit.NewDurableServer(redisConfig, redisClient)
	defer worker.Close()
	invite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, worker.URL())
	defer invite.Close()

	worker.RegisterInviteExpireHandler(invite.URL())

	var created struct {
		ID string `json:"id"`
	}
	postJSON(t, invite.URL()+"/v1/invites", map[string]any{
		"domain":         "party",
		"resource_id":    "party-1",
		"from_player_id": "p1",
		"to_player_id":   "p2",
		"ttl_seconds":    60,
	}, &created)

	result, err := worker.ExecuteUntilEmpty(context.Background(), "worker-a", "invite.expire", 10)
	if err != nil {
		t.Fatalf("execute until empty failed: %v", err)
	}
	if result.Completed != 1 {
		t.Fatalf("unexpected worker result: %+v", result)
	}

	var expired struct {
		Status string `json:"status"`
	}
	getJSON(t, invite.URL()+"/v1/invites/"+created.ID, &expired)
	if expired.Status != "expired" {
		t.Fatalf("expected durable invite to be expired, got %+v", expired)
	}
}

func TestLocalDurableChatWorkerOfflineDeliveryFlow(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	redisConfig, redisClient := newLocalRedisTestClient(t)

	worker := workertestkit.NewDurableServer(redisConfig, redisClient)
	defer worker.Close()
	chat := chattestkit.NewDurableServer(mysqlConfig, sqlDB, "", worker.URL())
	defer chat.Close()

	worker.RegisterChatOfflineDeliveryHandler(chat.URL())

	var conversation struct {
		ID string `json:"id"`
	}
	postJSON(t, chat.URL()+"/v1/conversations", map[string]any{
		"kind":              "private",
		"member_player_ids": []string{"p1", "p2"},
	}, &conversation)

	postJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/messages", map[string]any{
		"sender_player_id": "p1",
		"body":             "hello durable worker",
	}, nil)

	result, err := worker.ExecuteUntilEmpty(context.Background(), "worker-a", "chat.offline_delivery", 10)
	if err != nil {
		t.Fatalf("execute until empty failed: %v", err)
	}
	if result.Completed != 1 {
		t.Fatalf("unexpected worker result: %+v", result)
	}
	if chat.OfflineDeliveryCount(conversation.ID) != 1 {
		t.Fatalf("expected one durable offline delivery receipt")
	}
}

func TestLocalDurablePartyInviteJoinReadyFlow(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	invite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, "")
	defer invite.Close()

	redisConfig, redisClient := newLocalRedisTestClient(t)
	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()

	party := partytestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
	defer party.Close()

	var createdParty struct {
		ID string `json:"id"`
	}
	postJSON(t, party.URL()+"/v1/parties", map[string]any{
		"leader_id": "p1",
	}, &createdParty)

	var createdInvite struct {
		ID string `json:"id"`
	}
	postJSON(t, party.URL()+"/v1/parties/"+createdParty.ID+"/invites", map[string]any{
		"actor_player_id": "p1",
		"to_player_id":    "p2",
	}, &createdInvite)

	postJSON(t, invite.URL()+"/v1/invites/"+createdInvite.ID+"/accept", map[string]any{
		"actor_player_id": "p2",
	}, nil)

	postJSON(t, party.URL()+"/v1/parties/"+createdParty.ID+"/join", map[string]any{
		"invite_id":       createdInvite.ID,
		"actor_player_id": "p2",
	}, nil)

	postJSON(t, presence.URL()+"/v1/presence/connect", map[string]any{
		"player_id":  "p2",
		"session_id": "sess-2",
		"realm_id":   "realm-1",
		"location":   "lobby",
	}, nil)

	var readyState struct {
		IsReady bool `json:"is_ready"`
	}
	postJSON(t, party.URL()+"/v1/parties/"+createdParty.ID+"/ready", map[string]any{
		"actor_player_id": "p2",
		"is_ready":        true,
	}, &readyState)
	if !readyState.IsReady {
		t.Fatalf("expected durable party ready state to be true")
	}

	var members struct {
		Count   int `json:"count"`
		Members []struct {
			PlayerID string `json:"player_id"`
			Presence string `json:"presence"`
			IsReady  bool   `json:"is_ready"`
		} `json:"members"`
	}
	getJSON(t, party.URL()+"/v1/parties/"+createdParty.ID+"/members", &members)
	if members.Count != 2 {
		t.Fatalf("unexpected durable party members payload: %+v", members)
	}
}

func TestLocalDurableGuildInviteJoinFlow(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	invite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, "")
	defer invite.Close()

	redisConfig, redisClient := newLocalRedisTestClient(t)
	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()

	guild := guildtestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
	defer guild.Close()

	var createdGuild struct {
		ID string `json:"id"`
	}
	postJSON(t, guild.URL()+"/v1/guilds", map[string]any{
		"name":     "Durable Guild",
		"owner_id": "p1",
	}, &createdGuild)

	var createdInvite struct {
		ID string `json:"id"`
	}
	postJSON(t, guild.URL()+"/v1/guilds/"+createdGuild.ID+"/invites", map[string]any{
		"actor_player_id": "p1",
		"to_player_id":    "p2",
	}, &createdInvite)

	postJSON(t, invite.URL()+"/v1/invites/"+createdInvite.ID+"/accept", map[string]any{
		"actor_player_id": "p2",
	}, nil)

	postJSON(t, guild.URL()+"/v1/guilds/"+createdGuild.ID+"/join", map[string]any{
		"invite_id":       createdInvite.ID,
		"actor_player_id": "p2",
	}, nil)

	postJSON(t, presence.URL()+"/v1/presence/connect", map[string]any{
		"player_id":  "p2",
		"session_id": "sess-2",
		"realm_id":   "realm-1",
		"location":   "hall",
	}, nil)

	var members struct {
		Count   int `json:"count"`
		Members []struct {
			PlayerID string `json:"player_id"`
			Role     string `json:"role"`
			Presence string `json:"presence"`
		} `json:"members"`
	}
	getJSON(t, guild.URL()+"/v1/guilds/"+createdGuild.ID+"/members", &members)
	if members.Count != 2 {
		t.Fatalf("unexpected durable guild members payload: %+v", members)
	}
}

func TestLocalDurableGatewayRedisPersistsSessionAcrossRestart(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	identity := identitytestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer identity.Close()

	redisConfig, redisClient := newLocalRedisTestClient(t)
	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()

	gatewayA := gatewaytestkit.NewDurableServer(redisConfig, redisClient, identity.URL(), presence.URL(), "")

	var tokenPair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
		"account_id": "a1",
		"player_id":  "p1",
	}, &tokenPair)

	postJSON(t, gatewayA.URL()+"/v1/realtime/handshake", map[string]any{
		"access_token": tokenPair.AccessToken,
		"session_id":   "sess-1",
		"realm_id":     "realm-1",
		"location":     "lobby",
	}, nil)
	gatewayA.Close()

	gatewayB := gatewaytestkit.NewDurableServer(redisConfig, redisClient, identity.URL(), presence.URL(), "")
	defer gatewayB.Close()

	var session struct {
		SessionID string `json:"session_id"`
		PlayerID  string `json:"player_id"`
		State     string `json:"state"`
	}
	getJSON(t, gatewayB.URL()+"/v1/realtime/sessions/sess-1", &session)
	if session.SessionID != "sess-1" || session.PlayerID != "p1" || session.State != "active" {
		t.Fatalf("unexpected durable gateway session payload: %+v", session)
	}
}

func TestLocalDurablePresenceRedisPersistsAcrossRestart(t *testing.T) {
	requireLocalDurableTests(t)

	redisConfig, redisClient := newLocalRedisTestClient(t)

	presenceA := presencetestkit.NewDurableServer(redisConfig, redisClient)
	postJSON(t, presenceA.URL()+"/v1/presence/connect", map[string]any{
		"player_id":  "p1",
		"session_id": "sess-1",
		"realm_id":   "realm-1",
		"location":   "lobby",
	}, nil)
	presenceA.Close()

	presenceB := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presenceB.Close()

	var payload struct {
		PlayerID  string `json:"player_id"`
		Status    string `json:"status"`
		SessionID string `json:"session_id"`
	}
	getJSON(t, presenceB.URL()+"/v1/presence/p1", &payload)
	if payload.PlayerID != "p1" || payload.Status != "online" || payload.SessionID != "sess-1" {
		t.Fatalf("unexpected durable presence payload after restart: %+v", payload)
	}
}

func TestLocalDurableWorkerRedisPersistsQueuedJobsAcrossRestart(t *testing.T) {
	requireLocalDurableTests(t)

	redisConfig, redisClient := newLocalRedisTestClient(t)

	workerA := workertestkit.NewDurableServer(redisConfig, redisClient)
	postJSON(t, workerA.URL()+"/v1/jobs", map[string]any{
		"type":    "invite.expire",
		"payload": `{"invite_id":"inv-1"}`,
	}, nil)
	workerA.Close()

	workerB := workertestkit.NewDurableServer(redisConfig, redisClient)
	defer workerB.Close()

	var jobs struct {
		Count int `json:"count"`
		Jobs  []struct {
			Type string `json:"type"`
		} `json:"jobs"`
	}
	getJSON(t, workerB.URL()+"/v1/jobs", &jobs)
	if jobs.Count != 1 || len(jobs.Jobs) != 1 || jobs.Jobs[0].Type != "invite.expire" {
		t.Fatalf("unexpected durable worker payload after restart: %+v", jobs)
	}
}

func TestLocalDurableMySQLBootstrapRegistersMigrations(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	redisConfig, redisClient := newLocalRedisTestClient(t)
	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()

	startDurableMySQLBootstrappedServers := func() func() {
		identity := identitytestkit.NewDurableServer(mysqlConfig, sqlDB)
		social := socialtestkit.NewDurableServer(mysqlConfig, sqlDB)
		invite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, "")
		chat := chattestkit.NewDurableServer(mysqlConfig, sqlDB, "", "")
		party := partytestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
		guild := guildtestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
		return func() {
			guild.Close()
			party.Close()
			chat.Close()
			invite.Close()
			social.Close()
			identity.Close()
		}
	}

	closeFirstPass := startDurableMySQLBootstrappedServers()
	closeFirstPass()
	closeSecondPass := startDurableMySQLBootstrappedServers()
	closeSecondPass()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := sqlDB.QueryContext(ctx, `SELECT service_name, migration_id FROM schema_migrations ORDER BY service_name, migration_id`)
	if err != nil {
		t.Fatalf("query schema migrations: %v", err)
	}
	defer rows.Close()

	recorded := make(map[string][]string)
	for rows.Next() {
		var service string
		var migrationID string
		if err := rows.Scan(&service, &migrationID); err != nil {
			t.Fatalf("scan schema migration: %v", err)
		}
		recorded[service] = append(recorded[service], migrationID)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate schema migrations: %v", err)
	}

	expected := map[string][]string{
		"identity": {"001_identity_core"},
		"social":   {"001_social_core"},
		"invite":   {"001_invite_core"},
		"chat":     {"001_chat_core"},
		"party":    {"001_party_core", "002_party_queue", "003_party_assignment"},
		"guild":    {"001_guild_core", "002_guild_announcement", "003_guild_logs", "004_guild_progression"},
	}
	for service, migrationIDs := range expected {
		migrations := recorded[service]
		if len(migrations) != len(migrationIDs) {
			t.Fatalf("unexpected recorded migrations for %s: %+v", service, migrations)
		}
		for idx, migrationID := range migrationIDs {
			if migrations[idx] != migrationID {
				t.Fatalf("unexpected recorded migrations for %s: %+v", service, migrations)
			}
		}
	}
}

func TestLocalDurableOpsReadsMySQLBootstrapSnapshot(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	redisConfig, redisClient := newLocalRedisTestClient(t)
	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()

	identity := identitytestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer identity.Close()
	social := socialtestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer social.Close()
	invite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, "")
	defer invite.Close()
	chat := chattestkit.NewDurableServer(mysqlConfig, sqlDB, "", "")
	defer chat.Close()
	party := partytestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
	defer party.Close()
	guild := guildtestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
	defer guild.Close()

	ops := opstestkit.NewDurableServer(mysqlConfig, sqlDB, redisConfig, redisClient, presence.URL(), party.URL(), guild.URL(), "", social.URL())
	defer ops.Close()

	var snapshot struct {
		Count    int `json:"count"`
		Services []struct {
			Service      string   `json:"service"`
			Count        int      `json:"count"`
			MigrationIDs []string `json:"migration_ids"`
		} `json:"services"`
	}
	getJSON(t, ops.URL()+"/v1/ops/bootstrap/mysql", &snapshot)
	if snapshot.Count != 6 {
		t.Fatalf("unexpected mysql bootstrap snapshot count: %+v", snapshot)
	}
}

func TestLocalDurableOpsReadsRedisRuntimeSnapshot(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	redisConfig, redisClient := newLocalRedisTestClient(t)

	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()
	worker := workertestkit.NewDurableServer(redisConfig, redisClient)
	defer worker.Close()

	mysqlBackedInvite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, "")
	defer mysqlBackedInvite.Close()

	identity := identitytestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer identity.Close()
	gateway := gatewaytestkit.NewDurableServer(redisConfig, redisClient, identity.URL(), presence.URL(), "")
	defer gateway.Close()

	postJSON(t, presence.URL()+"/v1/presence/connect", map[string]any{
		"player_id":  "p1",
		"session_id": "sess-1",
		"realm_id":   "realm-1",
		"location":   "lobby",
	}, nil)
	postJSON(t, worker.URL()+"/v1/jobs", map[string]any{
		"type":    "invite.expire",
		"payload": `{"invite_id":"inv-1"}`,
	}, nil)

	var tokenPair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
		"account_id": "a1",
		"player_id":  "p1",
	}, &tokenPair)
	postJSON(t, gateway.URL()+"/v1/realtime/handshake", map[string]any{
		"access_token": tokenPair.AccessToken,
		"session_id":   "sess-1",
		"realm_id":     "realm-1",
		"location":     "lobby",
	}, nil)

	ops := opstestkit.NewDurableServer(mysqlConfig, sqlDB, redisConfig, redisClient, presence.URL(), "", "", worker.URL(), "")
	defer ops.Close()

	var snapshot struct {
		PresenceRecordCount  int `json:"presence_record_count"`
		GatewaySessionCount  int `json:"gateway_session_count"`
		WorkerJobCount       int `json:"worker_job_count"`
		WorkerStatusCounters []struct {
			Status string `json:"status"`
			Count  int    `json:"count"`
		} `json:"worker_status_counters"`
	}
	_ = mysqlBackedInvite
	getJSON(t, ops.URL()+"/v1/ops/runtime/redis", &snapshot)
	if snapshot.PresenceRecordCount != 1 || snapshot.GatewaySessionCount != 1 || snapshot.WorkerJobCount != 1 {
		t.Fatalf("unexpected redis runtime snapshot: %+v", snapshot)
	}
}

func TestLocalDurableOpsReadsDurableSummary(t *testing.T) {
	requireLocalDurableTests(t)

	mysqlConfig, sqlDB := newLocalMySQLTestDatabase(t)
	redisConfig, redisClient := newLocalRedisTestClient(t)

	presence := presencetestkit.NewDurableServer(redisConfig, redisClient)
	defer presence.Close()
	worker := workertestkit.NewDurableServer(redisConfig, redisClient)
	defer worker.Close()
	invite := invitetestkit.NewDurableServer(mysqlConfig, sqlDB, "")
	defer invite.Close()
	social := socialtestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer social.Close()
	party := partytestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
	defer party.Close()
	guild := guildtestkit.NewDurableServer(mysqlConfig, sqlDB, invite.URL(), presence.URL())
	defer guild.Close()
	identity := identitytestkit.NewDurableServer(mysqlConfig, sqlDB)
	defer identity.Close()
	gateway := gatewaytestkit.NewDurableServer(redisConfig, redisClient, identity.URL(), presence.URL(), "")
	defer gateway.Close()

	postJSON(t, presence.URL()+"/v1/presence/connect", map[string]any{
		"player_id":  "p1",
		"session_id": "sess-1",
		"realm_id":   "realm-1",
		"location":   "lobby",
	}, nil)
	postJSON(t, worker.URL()+"/v1/jobs", map[string]any{
		"type":    "invite.expire",
		"payload": `{"invite_id":"inv-1"}`,
	}, nil)

	var tokenPair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
		"account_id": "a1",
		"player_id":  "p1",
	}, &tokenPair)
	postJSON(t, gateway.URL()+"/v1/realtime/handshake", map[string]any{
		"access_token": tokenPair.AccessToken,
		"session_id":   "sess-1",
		"realm_id":     "realm-1",
		"location":     "lobby",
	}, nil)

	ops := opstestkit.NewDurableServer(mysqlConfig, sqlDB, redisConfig, redisClient, presence.URL(), party.URL(), guild.URL(), worker.URL(), social.URL())
	defer ops.Close()

	var summary struct {
		MySQL *struct {
			Count int `json:"count"`
		} `json:"mysql"`
		Redis *struct {
			PresenceRecordCount int `json:"presence_record_count"`
			GatewaySessionCount int `json:"gateway_session_count"`
			WorkerJobCount      int `json:"worker_job_count"`
		} `json:"redis"`
	}
	getJSON(t, ops.URL()+"/v1/ops/durable/summary", &summary)
	if summary.MySQL == nil || summary.MySQL.Count != 5 || summary.Redis == nil {
		t.Fatalf("unexpected durable summary: %+v", summary)
	}
	if summary.Redis.PresenceRecordCount != 1 || summary.Redis.GatewaySessionCount != 1 || summary.Redis.WorkerJobCount != 1 {
		t.Fatalf("unexpected redis summary payload: %+v", summary)
	}
}

func requireLocalDurableTests(t *testing.T) {
	t.Helper()
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("ENABLE_LOCAL_DURABLE_TESTS")), "true") {
		t.Skip("set ENABLE_LOCAL_DURABLE_TESTS=true to run local durable backend integration tests")
	}
}

func newLocalMySQLTestDatabase(t *testing.T) (db.MySQLConfig, *sql.DB) {
	t.Helper()

	baseConfig := db.LoadMySQLConfig()
	adminDB, err := sql.Open("mysql", adminDSN(baseConfig))
	if err != nil {
		t.Fatalf("open admin mysql connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := adminDB.PingContext(ctx); err != nil {
		_ = adminDB.Close()
		t.Fatalf("ping admin mysql connection: %v", err)
	}

	databaseName := fmt.Sprintf("social_backend_it_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE "+databaseName); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create temp mysql database: %v", err)
	}

	mysqlConfig := baseConfig
	mysqlConfig.Database = databaseName
	sqlDB, err := sql.Open("mysql", mysqlConfig.DSN())
	if err != nil {
		_, _ = adminDB.ExecContext(ctx, "DROP DATABASE "+databaseName)
		_ = adminDB.Close()
		t.Fatalf("open temp mysql database: %v", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		_, _ = adminDB.ExecContext(ctx, "DROP DATABASE "+databaseName)
		_ = adminDB.Close()
		t.Fatalf("ping temp mysql database: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dropCancel()
		_, _ = adminDB.ExecContext(dropCtx, "DROP DATABASE "+databaseName)
		_ = adminDB.Close()
	})

	return mysqlConfig, sqlDB
}

func newLocalRedisTestClient(t *testing.T) (db.RedisConfig, *redis.Client) {
	t.Helper()

	config := db.LoadRedisConfig()
	if os.Getenv("REDIS_DB") == "" {
		config.DB = 14 + rand.Intn(2)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("ping redis connection: %v", err)
	}
	if err := client.FlushDB(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("flush redis db: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_ = client.FlushDB(cleanupCtx).Err()
		_ = client.Close()
	})

	return config, client
}

func adminDSN(config db.MySQLConfig) string {
	adminConfig := config
	adminConfig.Database = ""
	dsn := adminConfig.DSN()
	return strings.Replace(dsn, "/?", "/mysql?", 1)
}
