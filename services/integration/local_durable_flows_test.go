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
	invitetestkit "github.com/xyun1996/social_backend/services/invite/testkit"
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
