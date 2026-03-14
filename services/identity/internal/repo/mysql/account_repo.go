package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/identity/internal/domain"
)

type schemaExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const (
	// AccountsTable is owned by identity for account to player mappings.
	AccountsTable = "identity_accounts"
	// RefreshTokensTable is owned by identity for durable refresh-token lineage.
	RefreshTokensTable = "identity_refresh_tokens"
)

// Repository is the MySQL-backed persistence implementation for identity.
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

// Migrations returns the versioned identity schema ownership.
func (r *Repository) Migrations() []db.Migration {
	return []db.Migration{
		{
			ID: "001_identity_core",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS identity_accounts (
					account_id VARCHAR(64) PRIMARY KEY,
					player_id VARCHAR(64) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);`,
				`CREATE TABLE IF NOT EXISTS identity_refresh_tokens (
					token_id VARCHAR(64) PRIMARY KEY,
					account_id VARCHAR(64) NOT NULL,
					player_id VARCHAR(64) NOT NULL,
					access_token VARCHAR(255) NOT NULL UNIQUE,
					refresh_token VARCHAR(255) NOT NULL,
					access_expires_at TIMESTAMP NOT NULL,
					refresh_expires_at TIMESTAMP NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					expires_at TIMESTAMP NULL,
					INDEX idx_identity_refresh_tokens_account (account_id)
				);`,
			},
		},
		{
			ID: "002_identity_refresh_expiry",
			Statements: []string{
				`ALTER TABLE identity_refresh_tokens
					ADD COLUMN refresh_expires_at TIMESTAMP NULL AFTER access_expires_at;`,
			},
		},
	}
}

// SchemaStatements returns the first-round identity schema ownership.
func (r *Repository) SchemaStatements() []string {
	return db.FlattenMigrations(r.Migrations())
}

// BootstrapSchema applies the identity-owned schema statements against the configured MySQL connection.
func (r *Repository) BootstrapSchema(ctx context.Context) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	return db.ApplyMySQLMigrations(ctx, r.sqlDB, "identity", r.Migrations())
}

// UpsertAccount persists the durable account-to-player mapping.
func (r *Repository) UpsertAccount(accountID string, playerID string) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO identity_accounts (account_id, player_id)
		 VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE player_id = VALUES(player_id)`,
		accountID,
		playerID,
	)
	return err
}

// SaveSession persists the latest issued token pair lineage for a player.
func (r *Repository) SaveSession(session domain.Session) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO identity_refresh_tokens (
			token_id,
			account_id,
			player_id,
			access_token,
			refresh_token,
			access_expires_at,
			refresh_expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			account_id = VALUES(account_id),
			player_id = VALUES(player_id),
			access_token = VALUES(access_token),
			access_expires_at = VALUES(access_expires_at),
			refresh_expires_at = VALUES(refresh_expires_at)`,
		session.RefreshToken,
		session.AccountID,
		session.PlayerID,
		session.AccessToken,
		session.RefreshToken,
		session.ExpiresAt.UTC(),
		session.RefreshExpiresAt.UTC(),
	)
	return err
}

// GetSessionByRefreshToken loads a persisted session by refresh token.
func (r *Repository) GetSessionByRefreshToken(refreshToken string) (domain.Session, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.Session{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT account_id, player_id, access_token, refresh_token, access_expires_at, refresh_expires_at
		 FROM identity_refresh_tokens
		 WHERE refresh_token = ?`,
		refreshToken,
	)

	var session domain.Session
	if err := row.Scan(&session.AccountID, &session.PlayerID, &session.AccessToken, &session.RefreshToken, &session.ExpiresAt, &session.RefreshExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, false, nil
		}
		return domain.Session{}, false, err
	}

	return session, true, nil
}

// GetSessionByAccessToken loads a persisted session by access token.
func (r *Repository) GetSessionByAccessToken(accessToken string) (domain.Session, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.Session{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT account_id, player_id, access_token, refresh_token, access_expires_at, refresh_expires_at
		 FROM identity_refresh_tokens
		 WHERE access_token = ?`,
		accessToken,
	)

	var session domain.Session
	if err := row.Scan(&session.AccountID, &session.PlayerID, &session.AccessToken, &session.RefreshToken, &session.ExpiresAt, &session.RefreshExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, false, nil
		}
		return domain.Session{}, false, err
	}

	return session, true, nil
}

// DeleteSessionByRefreshToken deletes a persisted session lineage by refresh token.
func (r *Repository) DeleteSessionByRefreshToken(refreshToken string) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`DELETE FROM identity_refresh_tokens WHERE refresh_token = ?`,
		refreshToken,
	)
	return err
}
