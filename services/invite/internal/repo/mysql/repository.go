package mysql

import (
	"database/sql"

	"github.com/xyun1996/social_backend/pkg/db"
)

const (
	// InvitesTable is owned by the invite service for durable lifecycle state.
	InvitesTable = "invites"
)

// Repository is the MySQL foundation for future invite persistence.
type Repository struct {
	config db.MySQLConfig
	sqlDB  *sql.DB
}

// NewRepository constructs the invite MySQL repository foundation.
func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository {
	return &Repository{config: config, sqlDB: sqlDB}
}

// DSN returns the shared MySQL DSN used by this repository.
func (r *Repository) DSN() string {
	return r.config.DSN()
}

// SchemaStatements returns the first-round invite schema ownership.
func (r *Repository) SchemaStatements() []string {
	return []string{
		`CREATE TABLE invites (
			invite_id VARCHAR(64) PRIMARY KEY,
			domain_name VARCHAR(32) NOT NULL,
			resource_id VARCHAR(64) NOT NULL,
			from_player_id VARCHAR(64) NOT NULL,
			to_player_id VARCHAR(64) NOT NULL,
			status VARCHAR(16) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			responded_at TIMESTAMP NULL,
			INDEX idx_invites_to_player (to_player_id, status),
			INDEX idx_invites_resource (domain_name, resource_id, status)
		);`,
	}
}
