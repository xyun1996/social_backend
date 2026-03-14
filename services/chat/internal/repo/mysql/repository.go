package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/chat/internal/domain"
)

type schemaExecutor interface { ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) }

const (
	ConversationsTable = "chat_conversations"
	ConversationMembersTable = "chat_conversation_members"
	MessagesTable = "chat_messages"
	ReadCursorsTable = "chat_read_cursors"
	ConversationGovernanceTable = "chat_conversation_governance"
)

type Repository struct { config db.MySQLConfig; sqlDB *sql.DB }
func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository { return &Repository{config: config, sqlDB: sqlDB} }
func (r *Repository) DSN() string { return r.config.DSN() }

func (r *Repository) Migrations() []db.Migration {
	return []db.Migration{
		{ID: "001_chat_core", Statements: []string{
			`CREATE TABLE IF NOT EXISTS chat_conversations (
				conversation_id VARCHAR(64) PRIMARY KEY,
				kind VARCHAR(32) NOT NULL,
				resource_id VARCHAR(64) NULL,
				send_policy VARCHAR(32) NOT NULL DEFAULT 'members',
				visibility_policy VARCHAR(32) NOT NULL DEFAULT 'members',
				moderation_mode VARCHAR(32) NOT NULL DEFAULT 'open',
				last_seq BIGINT NOT NULL DEFAULT 0,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				INDEX idx_chat_conversations_resource (kind, resource_id)
			);`,
			`CREATE TABLE IF NOT EXISTS chat_conversation_members (
				conversation_id VARCHAR(64) NOT NULL,
				player_id VARCHAR(64) NOT NULL,
				PRIMARY KEY (conversation_id, player_id),
				INDEX idx_chat_conversation_members_player (player_id)
			);`,
			`CREATE TABLE IF NOT EXISTS chat_messages (
				message_id VARCHAR(64) PRIMARY KEY,
				conversation_id VARCHAR(64) NOT NULL,
				seq BIGINT NOT NULL,
				sender_player_id VARCHAR(64) NOT NULL,
				body TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE KEY uq_chat_messages_conversation_seq (conversation_id, seq),
				INDEX idx_chat_messages_conversation_created (conversation_id, created_at)
			);`,
			`CREATE TABLE IF NOT EXISTS chat_read_cursors (
				conversation_id VARCHAR(64) NOT NULL,
				player_id VARCHAR(64) NOT NULL,
				ack_seq BIGINT NOT NULL DEFAULT 0,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (conversation_id, player_id),
				INDEX idx_chat_read_cursors_player (player_id)
			);`,
		}},
		{ID: "002_chat_governance", Statements: []string{
			`CREATE TABLE IF NOT EXISTS chat_conversation_governance (
				conversation_id VARCHAR(64) NOT NULL,
				player_id VARCHAR(64) NOT NULL,
				role VARCHAR(32) NOT NULL,
				PRIMARY KEY (conversation_id, player_id, role),
				INDEX idx_chat_conversation_governance_role (role)
			);`,
		}},
	}
}
func (r *Repository) SchemaStatements() []string { return db.FlattenMigrations(r.Migrations()) }
func (r *Repository) BootstrapSchema(ctx context.Context) error { if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }; return db.ApplyMySQLMigrations(ctx, r.sqlDB, "chat", r.Migrations()) }

func (r *Repository) ListConversations() ([]domain.Conversation, error) {
	if r == nil || r.sqlDB == nil { return nil, errors.New("mysql repository is not configured") }
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT conversation_id, kind, resource_id, send_policy, visibility_policy, moderation_mode, last_seq, created_at FROM chat_conversations`)
	if err != nil { return nil, err }
	defer rows.Close()
	conversations := make([]domain.Conversation, 0)
	for rows.Next() {
		conversation, err := r.scanConversation(rows)
		if err != nil { return nil, err }
		conversations = append(conversations, conversation)
	}
	return conversations, rows.Err()
}

func (r *Repository) SaveConversation(conversation domain.Conversation) error {
	if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }
	tx, err := r.sqlDB.BeginTx(context.Background(), nil)
	if err != nil { return err }
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(context.Background(), `INSERT INTO chat_conversations (conversation_id, kind, resource_id, send_policy, visibility_policy, moderation_mode, last_seq, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE kind = VALUES(kind), resource_id = VALUES(resource_id), send_policy = VALUES(send_policy), visibility_policy = VALUES(visibility_policy), moderation_mode = VALUES(moderation_mode), last_seq = VALUES(last_seq), created_at = VALUES(created_at)`,
		conversation.ID, conversation.Kind, nullString(conversation.ResourceID), conversation.SendPolicy, conversation.VisibilityPolicy, conversation.ModerationMode, conversation.LastSeq, conversation.CreatedAt.UTC()); err != nil { return err }
	if _, err := tx.ExecContext(context.Background(), `DELETE FROM chat_conversation_members WHERE conversation_id = ?`, conversation.ID); err != nil { return err }
	for _, memberID := range conversation.MemberPlayerIDs {
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO chat_conversation_members (conversation_id, player_id) VALUES (?, ?)`, conversation.ID, memberID); err != nil { return err }
	}
	if _, err := tx.ExecContext(context.Background(), `DELETE FROM chat_conversation_governance WHERE conversation_id = ?`, conversation.ID); err != nil { return err }
	for _, moderatorID := range conversation.ModeratorIDs {
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO chat_conversation_governance (conversation_id, player_id, role) VALUES (?, ?, 'moderator')`, conversation.ID, moderatorID); err != nil { return err }
	}
	for _, mutedID := range conversation.MutedPlayerIDs {
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO chat_conversation_governance (conversation_id, player_id, role) VALUES (?, ?, 'muted')`, conversation.ID, mutedID); err != nil { return err }
	}
	return tx.Commit()
}

func (r *Repository) GetConversation(conversationID string) (domain.Conversation, bool, error) {
	if r == nil || r.sqlDB == nil { return domain.Conversation{}, false, errors.New("mysql repository is not configured") }
	row := r.sqlDB.QueryRowContext(context.Background(), `SELECT conversation_id, kind, resource_id, send_policy, visibility_policy, moderation_mode, last_seq, created_at FROM chat_conversations WHERE conversation_id = ?`, conversationID)
	conversation, err := r.scanConversation(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return domain.Conversation{}, false, nil }
		return domain.Conversation{}, false, err
	}
	return conversation, true, nil
}

func (r *Repository) ListMessages(conversationID string) ([]domain.Message, error) {
	if r == nil || r.sqlDB == nil { return nil, errors.New("mysql repository is not configured") }
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT message_id, conversation_id, seq, sender_player_id, body, created_at FROM chat_messages WHERE conversation_id = ? ORDER BY seq ASC`, conversationID)
	if err != nil { return nil, err }
	defer rows.Close()
	messages := make([]domain.Message, 0)
	for rows.Next() { var message domain.Message; if err := rows.Scan(&message.ID, &message.ConversationID, &message.Seq, &message.SenderPlayerID, &message.Body, &message.CreatedAt); err != nil { return nil, err }; messages = append(messages, message) }
	return messages, rows.Err()
}

func (r *Repository) AppendMessage(message domain.Message) error {
	if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO chat_messages (message_id, conversation_id, seq, sender_player_id, body, created_at) VALUES (?, ?, ?, ?, ?, ?)`, message.ID, message.ConversationID, message.Seq, message.SenderPlayerID, message.Body, message.CreatedAt.UTC())
	return err
}

func (r *Repository) GetCursor(conversationID string, playerID string) (domain.ReadCursor, bool, error) {
	if r == nil || r.sqlDB == nil { return domain.ReadCursor{}, false, errors.New("mysql repository is not configured") }
	row := r.sqlDB.QueryRowContext(context.Background(), `SELECT conversation_id, player_id, ack_seq, updated_at FROM chat_read_cursors WHERE conversation_id = ? AND player_id = ?`, conversationID, playerID)
	var cursor domain.ReadCursor
	if err := row.Scan(&cursor.ConversationID, &cursor.PlayerID, &cursor.AckSeq, &cursor.UpdatedAt); err != nil { if errors.Is(err, sql.ErrNoRows) { return domain.ReadCursor{}, false, nil }; return domain.ReadCursor{}, false, err }
	return cursor, true, nil
}

func (r *Repository) SaveCursor(cursor domain.ReadCursor) error {
	if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO chat_read_cursors (conversation_id, player_id, ack_seq, updated_at) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE ack_seq = VALUES(ack_seq), updated_at = VALUES(updated_at)`, cursor.ConversationID, cursor.PlayerID, cursor.AckSeq, cursor.UpdatedAt.UTC())
	return err
}

func (r *Repository) scanConversation(scanner interface{ Scan(dest ...any) error }) (domain.Conversation, error) {
	var conversation domain.Conversation
	var resourceID sql.NullString
	if err := scanner.Scan(&conversation.ID, &conversation.Kind, &resourceID, &conversation.SendPolicy, &conversation.VisibilityPolicy, &conversation.ModerationMode, &conversation.LastSeq, &conversation.CreatedAt); err != nil { return domain.Conversation{}, err }
	if resourceID.Valid { conversation.ResourceID = resourceID.String }
	members, err := r.listMembers(conversation.ID)
	if err != nil { return domain.Conversation{}, err }
	conversation.MemberPlayerIDs = members
	moderators, muted, err := r.listGovernance(conversation.ID)
	if err != nil { return domain.Conversation{}, err }
	conversation.ModeratorIDs = moderators
	conversation.MutedPlayerIDs = muted
	return conversation, nil
}

func (r *Repository) listMembers(conversationID string) ([]string, error) {
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT player_id FROM chat_conversation_members WHERE conversation_id = ?`, conversationID)
	if err != nil { return nil, err }
	defer rows.Close()
	members := make([]string, 0)
	for rows.Next() { var playerID string; if err := rows.Scan(&playerID); err != nil { return nil, err }; members = append(members, playerID) }
	if err := rows.Err(); err != nil { return nil, err }
	slices.Sort(members)
	return members, nil
}

func (r *Repository) listGovernance(conversationID string) ([]string, []string, error) {
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT player_id, role FROM chat_conversation_governance WHERE conversation_id = ?`, conversationID)
	if err != nil { return nil, nil, err }
	defer rows.Close()
	moderators := make([]string, 0)
	muted := make([]string, 0)
	for rows.Next() {
		var playerID string
		var role string
		if err := rows.Scan(&playerID, &role); err != nil { return nil, nil, err }
		switch role {
		case "moderator": moderators = append(moderators, playerID)
		case "muted": muted = append(muted, playerID)
		}
	}
	if err := rows.Err(); err != nil { return nil, nil, err }
	slices.Sort(moderators)
	slices.Sort(muted)
	return moderators, muted, nil
}

func nullString(value string) any { if value == "" { return nil }; return value }
