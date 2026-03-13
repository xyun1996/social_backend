package db

import "testing"

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
