package db

import (
	"context"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestFlattenMigrationsPreservesOrder(t *testing.T) {
	t.Parallel()

	statements := FlattenMigrations([]Migration{
		{ID: "001", Statements: []string{"stmt-1", "stmt-2"}},
		{ID: "002", Statements: []string{"stmt-3"}},
	})

	if len(statements) != 3 || statements[0] != "stmt-1" || statements[1] != "stmt-2" || statements[2] != "stmt-3" {
		t.Fatalf("unexpected flattened statements: %+v", statements)
	}
}

func TestApplyMySQLMigrationsAppliesOnlyPendingSteps(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	migrations := []Migration{
		{ID: "001_init", Statements: []string{"stmt-1", "stmt-2"}},
		{ID: "002_extra", Statements: []string{"stmt-3"}},
	}

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("chat", "001_init").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("chat", "002_extra").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	mock.ExpectExec(regexp.QuoteMeta("stmt-3")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("chat", "002_extra").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := ApplyMySQLMigrations(context.Background(), sqlDB, "chat", migrations); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestApplyMySQLMigrationsStopsOnStatementError(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("invite", "001_init").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	mock.ExpectExec(regexp.QuoteMeta("stmt-1")).WillReturnError(errors.New("boom"))

	err = ApplyMySQLMigrations(context.Background(), sqlDB, "invite", []Migration{
		{ID: "001_init", Statements: []string{"stmt-1", "stmt-2"}},
	})
	if err == nil {
		t.Fatalf("expected migration application to fail")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
