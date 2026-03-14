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
	for _, statement := range repo.Migrations()[0].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("guild", "001_guild_core").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("guild", "002_guild_announcement").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.Migrations()[1].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("guild", "002_guild_announcement").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("guild", "003_guild_logs").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.Migrations()[2].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("guild", "003_guild_logs").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("guild", "004_guild_progression").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.Migrations()[3].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("guild", "004_guild_progression").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.BootstrapSchema(context.Background()); err != nil {
		t.Fatalf("BootstrapSchema returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySaveAndListLogs(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	entry := domain.GuildLogEntry{
		ID:        "log-1",
		GuildID:   "guild-1",
		Action:    "guild.created",
		ActorID:   "p1",
		TargetID:  "",
		Message:   "guild created",
		CreatedAt: time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO guild_logs (log_id, guild_id, action, actor_id, target_id, message, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   action = VALUES(action),
		   actor_id = VALUES(actor_id),
		   target_id = VALUES(target_id),
		   message = VALUES(message),
		   created_at = VALUES(created_at)`)).
		WithArgs(entry.ID, entry.GuildID, entry.Action, entry.ActorID, entry.TargetID, entry.Message, entry.CreatedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.SaveLog(entry); err != nil {
		t.Fatalf("SaveLog returned error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"log_id", "guild_id", "action", "actor_id", "target_id", "message", "created_at"}).
		AddRow(entry.ID, entry.GuildID, entry.Action, entry.ActorID, entry.TargetID, entry.Message, entry.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT log_id, guild_id, action, actor_id, target_id, message, created_at
		 FROM guild_logs
		 WHERE guild_id = ?`)).
		WithArgs(entry.GuildID).
		WillReturnRows(rows)

	logs, err := repo.ListLogs(entry.GuildID)
	if err != nil {
		t.Fatalf("ListLogs returned error: %v", err)
	}
	if len(logs) != 1 || logs[0].Action != entry.Action {
		t.Fatalf("unexpected guild logs: %+v", logs)
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
		ID:                    "guild-1",
		Name:                  "Guild",
		OwnerID:               "p1",
		Level:                 2,
		Experience:            140,
		Announcement:          "Welcome",
		AnnouncementUpdatedAt: time.Date(2026, 3, 13, 12, 5, 0, 0, time.UTC),
		Members: []domain.GuildMember{
			{PlayerID: "p1", Role: "owner", JoinedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)},
			{PlayerID: "p2", Role: "member", JoinedAt: time.Date(2026, 3, 13, 12, 1, 0, 0, time.UTC)},
		},
		CreatedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO guild_guilds (guild_id, name, owner_id, level, experience, announcement, announcement_updated_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   name = VALUES(name),
		   owner_id = VALUES(owner_id),
		   level = VALUES(level),
		   experience = VALUES(experience),
		   announcement = VALUES(announcement),
		   announcement_updated_at = VALUES(announcement_updated_at),
		   created_at = VALUES(created_at)`)).
		WithArgs(guild.ID, guild.Name, guild.OwnerID, guild.Level, guild.Experience, guild.Announcement, guild.AnnouncementUpdatedAt.UTC(), guild.CreatedAt.UTC()).
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

	row := sqlmock.NewRows([]string{"guild_id", "name", "owner_id", "level", "experience", "announcement", "announcement_updated_at", "created_at"}).
		AddRow(guild.ID, guild.Name, guild.OwnerID, guild.Level, guild.Experience, guild.Announcement, guild.AnnouncementUpdatedAt, guild.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT guild_id, name, owner_id, level, experience, announcement, announcement_updated_at, created_at FROM guild_guilds WHERE guild_id = ?`)).
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
	if !ok || len(loaded.Members) != 2 || loaded.Members[0].PlayerID != "p1" || loaded.Announcement != guild.Announcement || loaded.Level != guild.Level || loaded.Experience != guild.Experience {
		t.Fatalf("unexpected loaded guild: %+v", loaded)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySaveAndListActivities(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	record := domain.GuildActivityRecord{
		ID:          "act-1",
		GuildID:     "guild-1",
		TemplateKey: "donate",
		PlayerID:    "p1",
		DeltaXP:     25,
		CreatedAt:   time.Date(2026, 3, 14, 11, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO guild_activities (activity_id, guild_id, template_key, player_id, delta_xp, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   template_key = VALUES(template_key),
		   player_id = VALUES(player_id),
		   delta_xp = VALUES(delta_xp),
		   created_at = VALUES(created_at)`)).
		WithArgs(record.ID, record.GuildID, record.TemplateKey, record.PlayerID, record.DeltaXP, record.CreatedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.SaveActivity(record); err != nil {
		t.Fatalf("SaveActivity returned error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"activity_id", "guild_id", "template_key", "player_id", "delta_xp", "created_at"}).
		AddRow(record.ID, record.GuildID, record.TemplateKey, record.PlayerID, record.DeltaXP, record.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT activity_id, guild_id, template_key, player_id, delta_xp, created_at
		 FROM guild_activities
		 WHERE guild_id = ?`)).
		WithArgs(record.GuildID).
		WillReturnRows(rows)

	records, err := repo.ListActivities(record.GuildID)
	if err != nil {
		t.Fatalf("ListActivities returned error: %v", err)
	}
	if len(records) != 1 || records[0].TemplateKey != record.TemplateKey || records[0].DeltaXP != record.DeltaXP {
		t.Fatalf("unexpected activity records: %+v", records)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
