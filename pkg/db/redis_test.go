package db

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestRedisURL(t *testing.T) {
	t.Parallel()

	cfg := RedisConfig{
		Addr: "localhost:6379",
		DB:   0,
	}

	if got := cfg.URL(); got != "redis://localhost:6379/0" {
		t.Fatalf("unexpected redis url: %q", got)
	}
}

func TestOpenRedis(t *testing.T) {
	t.Parallel()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	defer server.Close()

	client, err := OpenRedis(context.Background(), RedisConfig{
		Addr: server.Addr(),
		DB:   0,
	})
	if err != nil {
		t.Fatalf("OpenRedis returned error: %v", err)
	}
	defer client.Close()

	if err := client.Set(context.Background(), "health", "ok", 0).Err(); err != nil {
		t.Fatalf("redis set failed: %v", err)
	}
}
