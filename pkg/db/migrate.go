package db

import (
	"context"
	"database/sql"
	"errors"
)

const schemaMigrationsTable = "schema_migrations"

// Migration defines a service-owned schema step.
type Migration struct {
	ID         string
	Statements []string
}

// FlattenMigrations returns the ordered statements for a migration list.
func FlattenMigrations(migrations []Migration) []string {
	statements := make([]string, 0)
	for _, migration := range migrations {
		statements = append(statements, migration.Statements...)
	}
	return statements
}

// ApplyMySQLMigrations applies unapplied service-owned migrations and records progress.
func ApplyMySQLMigrations(ctx context.Context, sqlDB *sql.DB, service string, migrations []Migration) error {
	if sqlDB == nil {
		return errors.New("mysql db is not configured")
	}
	if service == "" {
		return errors.New("migration service name is required")
	}

	if _, err := sqlDB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			service_name VARCHAR(64) NOT NULL,
			migration_id VARCHAR(128) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (service_name, migration_id)
		);`,
	); err != nil {
		return err
	}

	for _, migration := range migrations {
		applied, err := migrationApplied(ctx, sqlDB, service, migration.ID)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		for _, statement := range migration.Statements {
			if _, err := sqlDB.ExecContext(ctx, statement); err != nil {
				return err
			}
		}

		if _, err := sqlDB.ExecContext(
			ctx,
			`INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)`,
			service,
			migration.ID,
		); err != nil {
			return err
		}
	}

	return nil
}

func migrationApplied(ctx context.Context, sqlDB *sql.DB, service string, migrationID string) (bool, error) {
	row := sqlDB.QueryRowContext(
		ctx,
		`SELECT 1
		 FROM schema_migrations
		 WHERE service_name = ? AND migration_id = ?
		 LIMIT 1`,
		service,
		migrationID,
	)

	var marker int
	if err := row.Scan(&marker); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
