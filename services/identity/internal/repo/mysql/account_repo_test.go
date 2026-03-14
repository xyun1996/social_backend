package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/xyun1996/social_backend/pkg/db"
)

func TestBootstrapSchemaAppliesPendingMigration(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("identity", "001_identity_core").
		WillReturnError(sql.ErrNoRows)
	for _, statement := range repo.Migrations()[0].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("identity", "001_identity_core").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("identity", "002_identity_refresh_expiry").
		WillReturnError(sql.ErrNoRows)
	for _, statement := range repo.Migrations()[1].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("identity", "002_identity_refresh_expiry").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.BootstrapSchema(context.Background()); err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
