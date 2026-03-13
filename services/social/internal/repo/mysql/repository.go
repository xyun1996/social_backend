package mysql

import (
	"database/sql"

	"github.com/xyun1996/social_backend/pkg/db"
)

const (
	// FriendRequestsTable is owned by social for durable request lifecycle.
	FriendRequestsTable = "social_friend_requests"
	// FriendshipsTable is owned by social for accepted relationships.
	FriendshipsTable = "social_friendships"
	// BlocksTable is owned by social for durable block relationships.
	BlocksTable = "social_blocks"
)

// Repository is the MySQL foundation for future social persistence.
type Repository struct {
	config db.MySQLConfig
	sqlDB  *sql.DB
}

// NewRepository constructs the social MySQL repository foundation.
func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository {
	return &Repository{config: config, sqlDB: sqlDB}
}

// DSN returns the shared MySQL DSN used by this repository.
func (r *Repository) DSN() string {
	return r.config.DSN()
}

// SchemaStatements returns the first-round social schema ownership.
func (r *Repository) SchemaStatements() []string {
	return []string{
		`CREATE TABLE social_friend_requests (
			request_id VARCHAR(64) PRIMARY KEY,
			from_player_id VARCHAR(64) NOT NULL,
			to_player_id VARCHAR(64) NOT NULL,
			status VARCHAR(16) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			responded_at TIMESTAMP NULL,
			UNIQUE KEY uq_social_friend_requests_pair (from_player_id, to_player_id),
			INDEX idx_social_friend_requests_to_player (to_player_id, status)
		);`,
		`CREATE TABLE social_friendships (
			player_id VARCHAR(64) NOT NULL,
			friend_player_id VARCHAR(64) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (player_id, friend_player_id),
			INDEX idx_social_friendships_friend (friend_player_id)
		);`,
		`CREATE TABLE social_blocks (
			player_id VARCHAR(64) NOT NULL,
			blocked_player_id VARCHAR(64) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (player_id, blocked_player_id),
			INDEX idx_social_blocks_blocked (blocked_player_id)
		);`,
	}
}
