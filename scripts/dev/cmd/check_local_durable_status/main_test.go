package main

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("OPS_BASE_URL", "http://localhost:9088")
	t.Setenv("REQUIRE_MYSQL_SUMMARY", "true")
	t.Setenv("REQUIRE_REDIS_SUMMARY", "yes")
	t.Setenv("EXPECTED_MYSQL_SERVICES", "identity, social ,invite")

	cfg := loadConfig()
	if cfg.BaseURL != "http://localhost:9088" {
		t.Fatalf("unexpected base url: %+v", cfg)
	}
	if !cfg.RequireMySQLSummary || !cfg.RequireRedisSummary {
		t.Fatalf("unexpected required summary flags: %+v", cfg)
	}
	expectedServices := []string{"identity", "social", "invite"}
	if !reflect.DeepEqual(cfg.ExpectedMySQLServices, expectedServices) {
		t.Fatalf("unexpected expected mysql services: %+v", cfg.ExpectedMySQLServices)
	}
}

func TestValidateSummary(t *testing.T) {
	t.Parallel()

	summary := durableSummary{
		MySQL: &mysqlBootstrapSnapshot{
			Count: 2,
			Services: []mysqlBootstrapService{
				{Service: "identity", Count: 1},
				{Service: "invite", Count: 1},
			},
		},
		Redis: &redisRuntimeSnapshot{
			PresenceRecordCount: 1,
		},
	}

	problems := validateSummary(summary, config{
		RequireMySQLSummary:   true,
		RequireRedisSummary:   true,
		ExpectedMySQLServices: []string{"identity", "invite"},
	})
	if len(problems) != 0 {
		t.Fatalf("unexpected validation problems: %+v", problems)
	}
}

func TestValidateSummaryMissingExpectations(t *testing.T) {
	t.Parallel()

	problems := validateSummary(durableSummary{
		MySQL: &mysqlBootstrapSnapshot{
			Count: 1,
			Services: []mysqlBootstrapService{
				{Service: "identity", Count: 1},
			},
		},
	}, config{
		RequireMySQLSummary:   true,
		RequireRedisSummary:   true,
		ExpectedMySQLServices: []string{"identity", "chat"},
	})

	if len(problems) != 2 {
		t.Fatalf("unexpected validation problems: %+v", problems)
	}
	if problems[0] != "redis summary is required but missing" {
		t.Fatalf("unexpected first problem: %+v", problems)
	}
	if problems[1] != "mysql summary is missing expected services: chat" {
		t.Fatalf("unexpected second problem: %+v", problems)
	}
}
