package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/chat/internal/domain"
)

type schemaExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const (
	// ConversationsTable is owned by chat for durable conversation metadata.
	ConversationsTable = "chat_conversations"
	// ConversationMembersTable is owned by chat for durable member scope.
	ConversationMembersTable = "chat_conversation_members"
	// MessagesTable is owned by chat for durable ordered message history.
	MessagesTable = "chat_messages"
	// ReadCursorsTable is owned by chat for durable per-player ack state.
	ReadCursorsTable = "chat_read_cursors"
)

// Repository is the MySQL foundation for future chat persistence.
type Repository struct {
	config db.MySQLConfig
	sqlDB  *sql.DB
}

// NewRepository constructs the chat MySQL repository foundation.
func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository {
	return &Repository{config: config, sqlDB: sqlDB}
}

// DSN returns the shared MySQL DSN used by this repository.
func (r *Repository) DSN() string {
	return r.config.DSN()
}

// SchemaStatements returns the first-round chat schema ownership.
func (r *Repository) SchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS chat_conversations (
			conversation_id VARCHAR(64) PRIMARY KEY,
			kind VARCHAR(32) NOT NULL,
			resource_id VARCHAR(64) NULL,
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
	}
}

// BootstrapSchema applies the chat-owned schema statements against the configured MySQL connection.
func (r *Repository) BootstrapSchema(ctx context.Context) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	return applySchema(ctx, r.sqlDB, r.SchemaStatements())
}

// ListConversations returns all persisted conversations with their member scopes.
func (r *Repository) ListConversations() ([]domain.Conversation, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT conversation_id, kind, resource_id, last_seq, created_at FROM chat_conversations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]domain.Conversation, 0)
	for rows.Next() {
		var conversation domain.Conversation
		var resourceID sql.NullString
		if err := rows.Scan(&conversation.ID, &conversation.Kind, &resourceID, &conversation.LastSeq, &conversation.CreatedAt); err != nil {
			return nil, err
		}
		if resourceID.Valid {
			conversation.ResourceID = resourceID.String
		}

		members, err := r.listMembers(conversation.ID)
		if err != nil {
			return nil, err
		}
		slices.Sort(members)
		conversation.MemberPlayerIDs = members
		conversations = append(conversations, conversation)
	}
	return conversations, rows.Err()
}

// SaveConversation upserts conversation metadata and replaces its member scope.
func (r *Repository) SaveConversation(conversation domain.Conversation) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	tx, err := r.sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(
		context.Background(),
		`INSERT INTO chat_conversations (conversation_id, kind, resource_id, last_seq, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   kind = VALUES(kind),
		   resource_id = VALUES(resource_id),
		   last_seq = VALUES(last_seq),
		   created_at = VALUES(created_at)`,
		conversation.ID,
		conversation.Kind,
		nullString(conversation.ResourceID),
		conversation.LastSeq,
		conversation.CreatedAt.UTC(),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(), `DELETE FROM chat_conversation_members WHERE conversation_id = ?`, conversation.ID); err != nil {
		return err
	}
	for _, memberID := range conversation.MemberPlayerIDs {
		if _, err := tx.ExecContext(
			context.Background(),
			`INSERT INTO chat_conversation_members (conversation_id, player_id) VALUES (?, ?)`,
			conversation.ID,
			memberID,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetConversation loads a persisted conversation by id.
func (r *Repository) GetConversation(conversationID string) (domain.Conversation, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.Conversation{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT conversation_id, kind, resource_id, last_seq, created_at
		 FROM chat_conversations
		 WHERE conversation_id = ?`,
		conversationID,
	)

	var conversation domain.Conversation
	var resourceID sql.NullString
	if err := row.Scan(&conversation.ID, &conversation.Kind, &resourceID, &conversation.LastSeq, &conversation.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Conversation{}, false, nil
		}
		return domain.Conversation{}, false, err
	}
	if resourceID.Valid {
		conversation.ResourceID = resourceID.String
	}

	members, err := r.listMembers(conversationID)
	if err != nil {
		return domain.Conversation{}, false, err
	}
	slices.Sort(members)
	conversation.MemberPlayerIDs = members
	return conversation, true, nil
}

// ListMessages returns ordered messages for a conversation.
func (r *Repository) ListMessages(conversationID string) ([]domain.Message, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT message_id, conversation_id, seq, sender_player_id, body, created_at
		 FROM chat_messages
		 WHERE conversation_id = ?
		 ORDER BY seq ASC`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)
	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(&message.ID, &message.ConversationID, &message.Seq, &message.SenderPlayerID, &message.Body, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

// AppendMessage persists a new ordered chat message.
func (r *Repository) AppendMessage(message domain.Message) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO chat_messages (message_id, conversation_id, seq, sender_player_id, body, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		message.ID,
		message.ConversationID,
		message.Seq,
		message.SenderPlayerID,
		message.Body,
		message.CreatedAt.UTC(),
	)
	return err
}

// GetCursor loads a persisted read cursor.
func (r *Repository) GetCursor(conversationID string, playerID string) (domain.ReadCursor, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.ReadCursor{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT conversation_id, player_id, ack_seq, updated_at
		 FROM chat_read_cursors
		 WHERE conversation_id = ? AND player_id = ?`,
		conversationID,
		playerID,
	)

	var cursor domain.ReadCursor
	if err := row.Scan(&cursor.ConversationID, &cursor.PlayerID, &cursor.AckSeq, &cursor.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ReadCursor{}, false, nil
		}
		return domain.ReadCursor{}, false, err
	}
	return cursor, true, nil
}

// SaveCursor upserts a persisted read cursor.
func (r *Repository) SaveCursor(cursor domain.ReadCursor) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO chat_read_cursors (conversation_id, player_id, ack_seq, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   ack_seq = VALUES(ack_seq),
		   updated_at = VALUES(updated_at)`,
		cursor.ConversationID,
		cursor.PlayerID,
		cursor.AckSeq,
		cursor.UpdatedAt.UTC(),
	)
	return err
}

func (r *Repository) listMembers(conversationID string) ([]string, error) {
	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT player_id
		 FROM chat_conversation_members
		 WHERE conversation_id = ?`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]string, 0)
	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return nil, err
		}
		members = append(members, playerID)
	}
	return members, rows.Err()
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func applySchema(ctx context.Context, exec schemaExecutor, statements []string) error {
	for _, statement := range statements {
		if _, err := exec.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
