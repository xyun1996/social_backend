package mysql

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/chat/internal/domain"
)

func TestSaveAndGetConversation(t *testing.T) {
	t.Parallel()
	sqlDB, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock new: %v", err) }
	defer sqlDB.Close()
	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	createdAt := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)
	conversation := domain.Conversation{ID: "conv-1", Kind: "custom", ResourceID: "chan-1", MemberPlayerIDs: []string{"p1", "p2"}, SendPolicy: "moderated", VisibilityPolicy: "public_read", ModerationMode: "managed", ModeratorIDs: []string{"p1"}, MutedPlayerIDs: []string{"p2"}, LastSeq: 2, CreatedAt: createdAt}
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_conversations (conversation_id, kind, resource_id, send_policy, visibility_policy, moderation_mode, last_seq, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE kind = VALUES(kind), resource_id = VALUES(resource_id), send_policy = VALUES(send_policy), visibility_policy = VALUES(visibility_policy), moderation_mode = VALUES(moderation_mode), last_seq = VALUES(last_seq), created_at = VALUES(created_at)`)).WithArgs("conv-1", "custom", "chan-1", "moderated", "public_read", "managed", int64(2), createdAt.UTC()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM chat_conversation_members WHERE conversation_id = ?`)).WithArgs("conv-1").WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_conversation_members (conversation_id, player_id) VALUES (?, ?)`)).WithArgs("conv-1", "p1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_conversation_members (conversation_id, player_id) VALUES (?, ?)`)).WithArgs("conv-1", "p2").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM chat_conversation_governance WHERE conversation_id = ?`)).WithArgs("conv-1").WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_conversation_governance (conversation_id, player_id, role) VALUES (?, ?, 'moderator')`)).WithArgs("conv-1", "p1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_conversation_governance (conversation_id, player_id, role) VALUES (?, ?, 'muted')`)).WithArgs("conv-1", "p2").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := repo.SaveConversation(conversation); err != nil { t.Fatalf("save conversation: %v", err) }
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT conversation_id, kind, resource_id, send_policy, visibility_policy, moderation_mode, last_seq, created_at FROM chat_conversations WHERE conversation_id = ?`)).WithArgs("conv-1").WillReturnRows(sqlmock.NewRows([]string{"conversation_id", "kind", "resource_id", "send_policy", "visibility_policy", "moderation_mode", "last_seq", "created_at"}).AddRow("conv-1", "custom", "chan-1", "moderated", "public_read", "managed", int64(2), createdAt))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT player_id FROM chat_conversation_members WHERE conversation_id = ?`)).WithArgs("conv-1").WillReturnRows(sqlmock.NewRows([]string{"player_id"}).AddRow("p2").AddRow("p1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT player_id, role FROM chat_conversation_governance WHERE conversation_id = ?`)).WithArgs("conv-1").WillReturnRows(sqlmock.NewRows([]string{"player_id", "role"}).AddRow("p1", "moderator").AddRow("p2", "muted"))
	loaded, ok, err := repo.GetConversation("conv-1")
	if err != nil || !ok { t.Fatalf("unexpected get result: %+v ok=%v err=%v", loaded, ok, err) }
	if loaded.SendPolicy != "moderated" || len(loaded.ModeratorIDs) != 1 || len(loaded.MutedPlayerIDs) != 1 { t.Fatalf("unexpected loaded conversation: %+v", loaded) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("expectations: %v", err) }
}

func TestAppendMessageAndGetCursor(t *testing.T) {
	t.Parallel()
	sqlDB, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock new: %v", err) }
	defer sqlDB.Close()
	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	createdAt := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)
	message := domain.Message{ID: "msg-1", ConversationID: "conv-1", Seq: 1, SenderPlayerID: "p1", Body: "hello", CreatedAt: createdAt}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO chat_messages (message_id, conversation_id, seq, sender_player_id, body, created_at) VALUES (?, ?, ?, ?, ?, ?)")).WithArgs("msg-1", "conv-1", int64(1), "p1", "hello", createdAt.UTC()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.AppendMessage(message); err != nil { t.Fatalf("append message: %v", err) }
	cursorTime := createdAt.Add(time.Minute)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO chat_read_cursors (conversation_id, player_id, ack_seq, updated_at) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE ack_seq = VALUES(ack_seq), updated_at = VALUES(updated_at)")).WithArgs("conv-1", "p2", int64(1), cursorTime.UTC()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.SaveCursor(domain.ReadCursor{ConversationID: "conv-1", PlayerID: "p2", AckSeq: 1, UpdatedAt: cursorTime}); err != nil { t.Fatalf("save cursor: %v", err) }
	mock.ExpectQuery(regexp.QuoteMeta("SELECT conversation_id, player_id, ack_seq, updated_at FROM chat_read_cursors WHERE conversation_id = ? AND player_id = ?")).WithArgs("conv-1", "p2").WillReturnRows(sqlmock.NewRows([]string{"conversation_id", "player_id", "ack_seq", "updated_at"}).AddRow("conv-1", "p2", int64(1), cursorTime))
	cursor, ok, err := repo.GetCursor("conv-1", "p2")
	if err != nil || !ok || cursor.AckSeq != 1 { t.Fatalf("unexpected cursor: %+v ok=%v err=%v", cursor, ok, err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("expectations: %v", err) }
}

func TestBootstrapSchemaAppliesStatements(t *testing.T) {
	t.Parallel()
	sqlDB, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock new: %v", err) }
	defer sqlDB.Close()
	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).WillReturnResult(sqlmock.NewResult(0, 0))
	for _, migration := range repo.Migrations() {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).WithArgs("chat", migration.ID).WillReturnRows(sqlmock.NewRows([]string{"1"}))
		for _, statement := range migration.Statements { mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0)) }
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).WithArgs("chat", migration.ID).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	if err := repo.BootstrapSchema(context.Background()); err != nil { t.Fatalf("bootstrap schema: %v", err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("expectations: %v", err) }
}
