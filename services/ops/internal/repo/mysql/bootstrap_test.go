package mysql

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestGetMySQLBootstrapSnapshot(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	reader := NewBootstrapReader(sqlDB)
	rows := sqlmock.NewRows([]string{"service_name", "migration_id"}).
		AddRow("invite", "001_invite_core").
		AddRow("chat", "001_chat_core").
		AddRow("chat", "002_chat_indexes")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT service_name, migration_id
		 FROM schema_migrations
		 ORDER BY service_name, migration_id`)).
		WillReturnRows(rows)

	snapshot, appErr := reader.GetMySQLBootstrapSnapshot(context.Background())
	if appErr != nil {
		t.Fatalf("GetMySQLBootstrapSnapshot returned error: %+v", appErr)
	}
	if snapshot.Count != 2 {
		t.Fatalf("unexpected service count: %+v", snapshot)
	}
	if snapshot.Services[0].Service != "chat" || snapshot.Services[0].Count != 2 {
		t.Fatalf("unexpected chat service snapshot: %+v", snapshot.Services[0])
	}
	if snapshot.Services[1].Service != "invite" || snapshot.Services[1].Count != 1 {
		t.Fatalf("unexpected invite service snapshot: %+v", snapshot.Services[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
