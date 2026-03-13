package mysql

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/guild/internal/domain"
)

func TestRepositoryBootstrapSchema(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("guild", "001_guild_core").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.SchemaStatements() {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("guild", "001_guild_core").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.BootstrapSchema(context.Background()); err != nil {
		t.Fatalf("BootstrapSchema returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySaveAndLoadGuild(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	guild := domain.Guild{
		ID:      "guild-1",
		Name:    "Guild",
		OwnerID: "p1",
		Members: []domain.GuildMember{
			{PlayerID: "p1", Role: "owner", JoinedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)},
			{PlayerID: "p2", Role: "member", JoinedAt: time.Date(2026, 3, 13, 12, 1, 0, 0, time.UTC)},
		},
		CreatedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO guild_guilds (guild_id, name, owner_id, created_at)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   name = VALUES(name),
		   owner_id = VALUES(owner_id),
		   created_at = VALUES(created_at)`)).
		WithArgs(guild.ID, guild.Name, guild.OwnerID, guild.CreatedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM guild_members WHERE guild_id = ?`)).
		WithArgs(guild.ID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO guild_members (guild_id, player_id, role, joined_at) VALUES (?, ?, ?, ?)`)).
		WithArgs(guild.ID, "p1", "owner", guild.Members[0].JoinedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO guild_members (guild_id, player_id, role, joined_at) VALUES (?, ?, ?, ?)`)).
		WithArgs(guild.ID, "p2", "member", guild.Members[1].JoinedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.SaveGuild(guild); err != nil {
		t.Fatalf("SaveGuild returned error: %v", err)
	}

	row := sqlmock.NewRows([]string{"guild_id", "name", "owner_id", "created_at"}).
		AddRow(guild.ID, guild.Name, guild.OwnerID, guild.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT guild_id, name, owner_id, created_at FROM guild_guilds WHERE guild_id = ?`)).
		WithArgs(guild.ID).
		WillReturnRows(row)
	memberRows := sqlmock.NewRows([]string{"player_id", "role", "joined_at"}).
		AddRow("p2", "member", guild.Members[1].JoinedAt).
		AddRow("p1", "owner", guild.Members[0].JoinedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT player_id, role, joined_at FROM guild_members WHERE guild_id = ?`)).
		WithArgs(guild.ID).
		WillReturnRows(memberRows)

	loaded, ok, err := repo.GetGuild(guild.ID)
	if err != nil {
		t.Fatalf("GetGuild returned error: %v", err)
	}
	if !ok || len(loaded.Members) != 2 || loaded.Members[0].PlayerID != "p1" {
		t.Fatalf("unexpected loaded guild: %+v", loaded)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
