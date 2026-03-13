package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/invite/internal/domain"
)

type schemaExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

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

// BootstrapSchema applies the invite-owned schema statements against the configured MySQL connection.
func (r *Repository) BootstrapSchema(ctx context.Context) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	return applySchema(ctx, r.sqlDB, r.SchemaStatements())
}

// ListInvites returns all persisted invites ordered by created time then id.
func (r *Repository) ListInvites() ([]domain.Invite, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT invite_id, domain_name, resource_id, from_player_id, to_player_id, status, created_at, expires_at, responded_at
		 FROM invites`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invites := make([]domain.Invite, 0)
	for rows.Next() {
		var invite domain.Invite
		var respondedAt sql.NullTime
		if err := rows.Scan(
			&invite.ID,
			&invite.Domain,
			&invite.ResourceID,
			&invite.FromPlayerID,
			&invite.ToPlayerID,
			&invite.Status,
			&invite.CreatedAt,
			&invite.ExpiresAt,
			&respondedAt,
		); err != nil {
			return nil, err
		}
		if respondedAt.Valid {
			value := respondedAt.Time
			invite.RespondedAt = &value
		}
		invites = append(invites, invite)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(invites, func(a domain.Invite, b domain.Invite) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})
	return invites, nil
}

// SaveInvite upserts a persisted invite lifecycle record.
func (r *Repository) SaveInvite(invite domain.Invite) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO invites (
			invite_id,
			domain_name,
			resource_id,
			from_player_id,
			to_player_id,
			status,
			created_at,
			expires_at,
			responded_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			domain_name = VALUES(domain_name),
			resource_id = VALUES(resource_id),
			from_player_id = VALUES(from_player_id),
			to_player_id = VALUES(to_player_id),
			status = VALUES(status),
			created_at = VALUES(created_at),
			expires_at = VALUES(expires_at),
			responded_at = VALUES(responded_at)`,
		invite.ID,
		invite.Domain,
		invite.ResourceID,
		invite.FromPlayerID,
		invite.ToPlayerID,
		invite.Status,
		invite.CreatedAt.UTC(),
		invite.ExpiresAt.UTC(),
		nullTime(invite.RespondedAt),
	)
	return err
}

// GetInvite loads a persisted invite by id.
func (r *Repository) GetInvite(inviteID string) (domain.Invite, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.Invite{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT invite_id, domain_name, resource_id, from_player_id, to_player_id, status, created_at, expires_at, responded_at
		 FROM invites
		 WHERE invite_id = ?`,
		inviteID,
	)

	var invite domain.Invite
	var respondedAt sql.NullTime
	if err := row.Scan(
		&invite.ID,
		&invite.Domain,
		&invite.ResourceID,
		&invite.FromPlayerID,
		&invite.ToPlayerID,
		&invite.Status,
		&invite.CreatedAt,
		&invite.ExpiresAt,
		&respondedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Invite{}, false, nil
		}
		return domain.Invite{}, false, err
	}
	if respondedAt.Valid {
		value := respondedAt.Time
		invite.RespondedAt = &value
	}
	return invite, true, nil
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func applySchema(ctx context.Context, exec schemaExecutor, statements []string) error {
	for _, statement := range statements {
		if _, err := exec.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
