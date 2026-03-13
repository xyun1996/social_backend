package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

type config struct {
	BaseURL               string
	RequireMySQLSummary   bool
	RequireRedisSummary   bool
	ExpectedMySQLServices []string
}

type durableSummary struct {
	MySQL *mysqlBootstrapSnapshot `json:"mysql"`
	Redis *redisRuntimeSnapshot   `json:"redis"`
}

type mysqlBootstrapSnapshot struct {
	Count    int                     `json:"count"`
	Services []mysqlBootstrapService `json:"services"`
}

type mysqlBootstrapService struct {
	Service      string   `json:"service"`
	Count        int      `json:"count"`
	MigrationIDs []string `json:"migration_ids"`
}

type redisRuntimeSnapshot struct {
	RedisURL             string                   `json:"redis_url"`
	PresenceRecordCount  int                      `json:"presence_record_count"`
	GatewaySessionCount  int                      `json:"gateway_session_count"`
	WorkerJobCount       int                      `json:"worker_job_count"`
	WorkerStatusCounters []redisWorkerStatusCount `json:"worker_status_counters"`
}

type redisWorkerStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func main() {
	cfg := loadConfig()
	summary := fetchSummary(cfg.BaseURL)
	printJSON("Durable summary", summary)

	problems := validateSummary(summary, cfg)
	if len(problems) == 0 {
		fmt.Println("Durable status checks passed.")
		return
	}

	for _, problem := range problems {
		fmt.Fprintf(os.Stderr, "durable status check failed: %s\n", problem)
	}
	os.Exit(1)
}

func loadConfig() config {
	baseURL := strings.TrimSpace(os.Getenv("OPS_BASE_URL"))
	if baseURL == "" {
		baseURL = "http://localhost:8088"
	}

	return config{
		BaseURL:               baseURL,
		RequireMySQLSummary:   parseBoolEnv("REQUIRE_MYSQL_SUMMARY"),
		RequireRedisSummary:   parseBoolEnv("REQUIRE_REDIS_SUMMARY"),
		ExpectedMySQLServices: parseCSVEnv("EXPECTED_MYSQL_SERVICES"),
	}
}

func fetchSummary(baseURL string) durableSummary {
	url := strings.TrimRight(baseURL, "/") + "/v1/ops/durable/summary"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("GET %s returned status %d", url, resp.StatusCode)
	}

	var payload durableSummary
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		log.Fatalf("decode %s failed: %v", url, err)
	}
	return payload
}

func validateSummary(summary durableSummary, cfg config) []string {
	problems := make([]string, 0)

	if cfg.RequireMySQLSummary && summary.MySQL == nil {
		problems = append(problems, "mysql summary is required but missing")
	}
	if cfg.RequireRedisSummary && summary.Redis == nil {
		problems = append(problems, "redis summary is required but missing")
	}

	if summary.MySQL != nil && len(cfg.ExpectedMySQLServices) > 0 {
		recorded := map[string]struct{}{}
		for _, service := range summary.MySQL.Services {
			recorded[service.Service] = struct{}{}
		}

		missing := make([]string, 0)
		for _, service := range cfg.ExpectedMySQLServices {
			if _, ok := recorded[service]; !ok {
				missing = append(missing, service)
			}
		}
		if len(missing) > 0 {
			slices.Sort(missing)
			problems = append(problems, "mysql summary is missing expected services: "+strings.Join(missing, ", "))
		}
	}

	return problems
}

func parseBoolEnv(key string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseCSVEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	items := strings.Split(raw, ",")
	parsed := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		parsed = append(parsed, value)
	}
	return parsed
}

func printJSON(label string, payload any) {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Fatalf("marshal %s failed: %v", label, err)
	}
	fmt.Printf("%s\n%s\n", label, string(raw))
}
