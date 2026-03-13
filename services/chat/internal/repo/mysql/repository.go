package mysql

import (
	"database/sql"

	"github.com/xyun1996/social_backend/pkg/db"
)

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
		`CREATE TABLE chat_conversations (
			conversation_id VARCHAR(64) PRIMARY KEY,
			kind VARCHAR(32) NOT NULL,
			resource_id VARCHAR(64) NULL,
			last_seq BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_chat_conversations_resource (kind, resource_id)
		);`,
		`CREATE TABLE chat_conversation_members (
			conversation_id VARCHAR(64) NOT NULL,
			player_id VARCHAR(64) NOT NULL,
			PRIMARY KEY (conversation_id, player_id),
			INDEX idx_chat_conversation_members_player (player_id)
		);`,
		`CREATE TABLE chat_messages (
			message_id VARCHAR(64) PRIMARY KEY,
			conversation_id VARCHAR(64) NOT NULL,
			seq BIGINT NOT NULL,
			sender_player_id VARCHAR(64) NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uq_chat_messages_conversation_seq (conversation_id, seq),
			INDEX idx_chat_messages_conversation_created (conversation_id, created_at)
		);`,
		`CREATE TABLE chat_read_cursors (
			conversation_id VARCHAR(64) NOT NULL,
			player_id VARCHAR(64) NOT NULL,
			ack_seq BIGINT NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (conversation_id, player_id),
			INDEX idx_chat_read_cursors_player (player_id)
		);`,
	}
}
