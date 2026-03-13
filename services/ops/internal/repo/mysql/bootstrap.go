package mysql

import (
	"context"
	"database/sql"
	"slices"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

// BootstrapReader reads operator-facing MySQL migration status.
type BootstrapReader struct {
	sqlDB *sql.DB
}

// NewBootstrapReader constructs the ops MySQL bootstrap reader.
func NewBootstrapReader(sqlDB *sql.DB) *BootstrapReader {
	return &BootstrapReader{sqlDB: sqlDB}
}

// GetMySQLBootstrapSnapshot reads the current schema_migrations state.
func (r *BootstrapReader) GetMySQLBootstrapSnapshot(ctx context.Context) (opsservice.MySQLBootstrapSnapshot, *apperrors.Error) {
	if r == nil || r.sqlDB == nil {
		err := apperrors.New("dependency_missing", "mysql bootstrap reader is not configured", 500)
		return opsservice.MySQLBootstrapSnapshot{}, &err
	}

	rows, err := r.sqlDB.QueryContext(
		ctx,
		`SELECT service_name, migration_id
		 FROM schema_migrations
		 ORDER BY service_name, migration_id`,
	)
	if err != nil {
		appErr := apperrors.New("db_query_failed", err.Error(), 500)
		return opsservice.MySQLBootstrapSnapshot{}, &appErr
	}
	defer rows.Close()

	services := map[string][]string{}
	for rows.Next() {
		var service string
		var migrationID string
		if err := rows.Scan(&service, &migrationID); err != nil {
			appErr := apperrors.New("db_query_failed", err.Error(), 500)
			return opsservice.MySQLBootstrapSnapshot{}, &appErr
		}
		services[service] = append(services[service], migrationID)
	}
	if err := rows.Err(); err != nil {
		appErr := apperrors.New("db_query_failed", err.Error(), 500)
		return opsservice.MySQLBootstrapSnapshot{}, &appErr
	}

	names := make([]string, 0, len(services))
	for service := range services {
		names = append(names, service)
	}
	slices.Sort(names)

	snapshot := opsservice.MySQLBootstrapSnapshot{
		Count:    len(names),
		Services: make([]opsservice.MySQLBootstrapService, 0, len(names)),
	}
	for _, service := range names {
		migrationIDs := append([]string(nil), services[service]...)
		slices.Sort(migrationIDs)
		snapshot.Services = append(snapshot.Services, opsservice.MySQLBootstrapService{
			Service:      service,
			Count:        len(migrationIDs),
			MigrationIDs: migrationIDs,
		})
	}
	return snapshot, nil
}
