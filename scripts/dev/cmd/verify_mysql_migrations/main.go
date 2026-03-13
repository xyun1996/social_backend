package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xyun1996/social_backend/pkg/db"
)

var expectedServiceMigrations = map[string][]string{
	"identity": {"001_identity_core"},
	"social":   {"001_social_core"},
	"invite":   {"001_invite_core"},
	"chat":     {"001_chat_core"},
	"party":    {"001_party_core"},
	"guild":    {"001_guild_core"},
}

func main() {
	mysqlConfig := db.LoadMySQLConfig()
	sqlDB, err := sql.Open("mysql", mysqlConfig.DSN())
	if err != nil {
		log.Fatalf("open mysql connection: %v", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("ping mysql connection: %v", err)
	}

	applied, err := loadAppliedMigrations(ctx, sqlDB)
	if err != nil {
		log.Fatalf("load schema migrations: %v", err)
	}

	for service, migrationIDs := range expectedServiceMigrations {
		appliedIDs := applied[service]
		for _, migrationID := range migrationIDs {
			if !contains(appliedIDs, migrationID) {
				log.Fatalf("missing schema migration %s/%s", service, migrationID)
			}
		}
	}

	services := make([]string, 0, len(applied))
	for service := range applied {
		services = append(services, service)
	}
	sort.Strings(services)

	for _, service := range services {
		ids := append([]string(nil), applied[service]...)
		sort.Strings(ids)
		fmt.Printf("%s: %v\n", service, ids)
	}
}

func loadAppliedMigrations(ctx context.Context, sqlDB *sql.DB) (map[string][]string, error) {
	rows, err := sqlDB.QueryContext(
		ctx,
		`SELECT service_name, migration_id
		 FROM schema_migrations`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string][]string)
	for rows.Next() {
		var service string
		var migrationID string
		if err := rows.Scan(&service, &migrationID); err != nil {
			return nil, err
		}
		applied[service] = append(applied[service], migrationID)
	}
	return applied, rows.Err()
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
