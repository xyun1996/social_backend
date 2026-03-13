package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/party/internal/domain"
)

type schemaExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const (
	PartiesTable      = "party_parties"
	PartyMembersTable = "party_members"
	PartyReadyTable   = "party_ready_states"
)

// Repository implements party durable storage on MySQL.
type Repository struct {
	config db.MySQLConfig
	sqlDB  *sql.DB
}

func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository {
	return &Repository{config: config, sqlDB: sqlDB}
}

func (r *Repository) DSN() string {
	return r.config.DSN()
}

// Migrations returns the versioned party schema ownership.
func (r *Repository) Migrations() []db.Migration {
	return []db.Migration{
		{
			ID: "001_party_core",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS party_parties (
					party_id VARCHAR(64) PRIMARY KEY,
					leader_id VARCHAR(64) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);`,
				`CREATE TABLE IF NOT EXISTS party_members (
					party_id VARCHAR(64) NOT NULL,
					player_id VARCHAR(64) NOT NULL,
					PRIMARY KEY (party_id, player_id),
					INDEX idx_party_members_player (player_id)
				);`,
				`CREATE TABLE IF NOT EXISTS party_ready_states (
					party_id VARCHAR(64) NOT NULL,
					player_id VARCHAR(64) NOT NULL,
					is_ready BOOLEAN NOT NULL,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (party_id, player_id)
				);`,
			},
		},
	}
}

func (r *Repository) SchemaStatements() []string {
	return db.FlattenMigrations(r.Migrations())
}

func (r *Repository) BootstrapSchema(ctx context.Context) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	return db.ApplyMySQLMigrations(ctx, r.sqlDB, "party", r.Migrations())
}

func (r *Repository) SaveParty(party domain.Party) error {
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
		`INSERT INTO party_parties (party_id, leader_id, created_at)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   leader_id = VALUES(leader_id),
		   created_at = VALUES(created_at)`,
		party.ID,
		party.LeaderID,
		party.CreatedAt.UTC(),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(), `DELETE FROM party_members WHERE party_id = ?`, party.ID); err != nil {
		return err
	}
	for _, memberID := range party.MemberIDs {
		if _, err := tx.ExecContext(
			context.Background(),
			`INSERT INTO party_members (party_id, player_id) VALUES (?, ?)`,
			party.ID,
			memberID,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetParty(partyID string) (domain.Party, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.Party{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT party_id, leader_id, created_at FROM party_parties WHERE party_id = ?`,
		partyID,
	)

	var party domain.Party
	if err := row.Scan(&party.ID, &party.LeaderID, &party.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Party{}, false, nil
		}
		return domain.Party{}, false, err
	}

	memberIDs, err := r.listMembers(partyID)
	if err != nil {
		return domain.Party{}, false, err
	}
	slices.Sort(memberIDs)
	party.MemberIDs = memberIDs
	return party, true, nil
}

func (r *Repository) ListParties() ([]domain.Party, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT party_id, leader_id, created_at FROM party_parties`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parties := make([]domain.Party, 0)
	for rows.Next() {
		var party domain.Party
		if err := rows.Scan(&party.ID, &party.LeaderID, &party.CreatedAt); err != nil {
			return nil, err
		}
		memberIDs, err := r.listMembers(party.ID)
		if err != nil {
			return nil, err
		}
		slices.Sort(memberIDs)
		party.MemberIDs = memberIDs
		parties = append(parties, party)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(parties, func(a domain.Party, b domain.Party) int {
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
	return parties, nil
}

func (r *Repository) SaveReadyState(state domain.ReadyState) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO party_ready_states (party_id, player_id, is_ready, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   is_ready = VALUES(is_ready),
		   updated_at = VALUES(updated_at)`,
		state.PartyID,
		state.PlayerID,
		state.IsReady,
		state.UpdatedAt.UTC(),
	)
	return err
}

func (r *Repository) ListReadyStates(partyID string) ([]domain.ReadyState, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT party_id, player_id, is_ready, updated_at
		 FROM party_ready_states
		 WHERE party_id = ?`,
		partyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make([]domain.ReadyState, 0)
	for rows.Next() {
		var state domain.ReadyState
		if err := rows.Scan(&state.PartyID, &state.PlayerID, &state.IsReady, &state.UpdatedAt); err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(states, func(a domain.ReadyState, b domain.ReadyState) int {
		switch {
		case a.PlayerID < b.PlayerID:
			return -1
		case a.PlayerID > b.PlayerID:
			return 1
		default:
			return 0
		}
	})
	return states, nil
}

func (r *Repository) listMembers(partyID string) ([]string, error) {
	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT player_id FROM party_members WHERE party_id = ?`,
		partyID,
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
