package mysql

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/invite/internal/domain"
)

func TestSaveAndGetInvite(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	createdAt := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Minute)
	respondedAt := createdAt.Add(30 * time.Second)
	invite := domain.Invite{
		ID:           "inv-1",
		Domain:       "party",
		ResourceID:   "party-1",
		FromPlayerID: "p1",
		ToPlayerID:   "p2",
		Status:       "accepted",
		CreatedAt:    createdAt,
		ExpiresAt:    expiresAt,
		RespondedAt:  &respondedAt,
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO invites (")).
		WithArgs("inv-1", "party", "party-1", "p1", "p2", "accepted", createdAt.UTC(), expiresAt.UTC(), respondedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.SaveInvite(invite); err != nil {
		t.Fatalf("save invite: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT invite_id, domain_name, resource_id, from_player_id, to_player_id, status, created_at, expires_at, responded_at FROM invites WHERE invite_id = ?")).
		WithArgs("inv-1").
		WillReturnRows(sqlmock.NewRows([]string{"invite_id", "domain_name", "resource_id", "from_player_id", "to_player_id", "status", "created_at", "expires_at", "responded_at"}).
			AddRow("inv-1", "party", "party-1", "p1", "p2", "accepted", createdAt, expiresAt, respondedAt))

	loaded, ok, err := repo.GetInvite("inv-1")
	if err != nil {
		t.Fatalf("get invite: %v", err)
	}
	if !ok || loaded.Status != "accepted" || loaded.RespondedAt == nil {
		t.Fatalf("unexpected loaded invite: %+v ok=%v", loaded, ok)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestListInvitesOrdersByCreatedAtThenID(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	base := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT invite_id, domain_name, resource_id, from_player_id, to_player_id, status, created_at, expires_at, responded_at FROM invites")).
		WillReturnRows(sqlmock.NewRows([]string{"invite_id", "domain_name", "resource_id", "from_player_id", "to_player_id", "status", "created_at", "expires_at", "responded_at"}).
			AddRow("inv-2", "party", "party-1", "p1", "p3", "pending", base.Add(time.Minute), base.Add(2*time.Minute), nil).
			AddRow("inv-1", "party", "party-1", "p1", "p2", "pending", base, base.Add(2*time.Minute), nil))

	invites, err := repo.ListInvites()
	if err != nil {
		t.Fatalf("list invites: %v", err)
	}
	if len(invites) != 2 || invites[0].ID != "inv-1" || invites[1].ID != "inv-2" {
		t.Fatalf("unexpected invite ordering: %+v", invites)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestBootstrapSchemaAppliesStatements(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("invite", "001_invite_core").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.SchemaStatements() {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("invite", "001_invite_core").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.BootstrapSchema(context.Background()); err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
