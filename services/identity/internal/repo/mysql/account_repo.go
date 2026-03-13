package mysql

import (
	"database/sql"

	"github.com/xyun1996/social_backend/pkg/db"
)

const (
	// AccountsTable is owned by identity for account to player mappings.
	AccountsTable = "identity_accounts"
	// RefreshTokensTable is owned by identity for durable refresh-token lineage.
	RefreshTokensTable = "identity_refresh_tokens"
)

// Repository is the MySQL foundation for future identity persistence.
type Repository struct {
	config db.MySQLConfig
	sqlDB  *sql.DB
}

// NewRepository constructs the identity MySQL repository foundation.
func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository {
	return &Repository{config: config, sqlDB: sqlDB}
}

// DSN returns the shared MySQL DSN used by this repository.
func (r *Repository) DSN() string {
	return r.config.DSN()
}

// SchemaStatements returns the first-round identity schema ownership.
func (r *Repository) SchemaStatements() []string {
	return []string{
		`CREATE TABLE identity_accounts (
			account_id VARCHAR(64) PRIMARY KEY,
			player_id VARCHAR(64) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE identity_refresh_tokens (
			token_id VARCHAR(64) PRIMARY KEY,
			account_id VARCHAR(64) NOT NULL,
			player_id VARCHAR(64) NOT NULL,
			refresh_token VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NULL,
			INDEX idx_identity_refresh_tokens_account (account_id)
		);`,
	}
}
