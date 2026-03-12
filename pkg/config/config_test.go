package config

import "testing"

func TestLoadServiceConfigUsesDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("GATEWAY_ADDR", "")

	cfg := LoadServiceConfig("gateway", ":8080")

	if cfg.Name != "gateway" {
		t.Fatalf("unexpected name: got %q", cfg.Name)
	}

	if cfg.Env != "local" {
		t.Fatalf("unexpected env: got %q want %q", cfg.Env, "local")
	}

	if cfg.Addr != ":8080" {
		t.Fatalf("unexpected addr: got %q want %q", cfg.Addr, ":8080")
	}
}

func TestLoadServiceConfigUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("IDENTITY_ADDR", ":9091")

	cfg := LoadServiceConfig("identity", ":8081")

	if cfg.Env != "dev" {
		t.Fatalf("unexpected env: got %q want %q", cfg.Env, "dev")
	}

	if cfg.Addr != ":9091" {
		t.Fatalf("unexpected addr: got %q want %q", cfg.Addr, ":9091")
	}
}
