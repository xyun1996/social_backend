package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/social/internal/domain"
)

type schemaExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const (
	// FriendRequestsTable is owned by social for durable request lifecycle.
	FriendRequestsTable = "social_friend_requests"
	// FriendshipsTable is owned by social for accepted relationships.
	FriendshipsTable = "social_friendships"
	// BlocksTable is owned by social for durable block relationships.
	BlocksTable = "social_blocks"
	// FriendRemarksTable stores player-defined friend metadata.
	FriendRemarksTable = "social_friend_remarks"
)

// Repository is the MySQL repository for durable social state.
type Repository struct {
	config db.MySQLConfig
	sqlDB  *sql.DB
}

// NewRepository constructs the social MySQL repository.
func NewRepository(config db.MySQLConfig, sqlDB *sql.DB) *Repository {
	return &Repository{config: config, sqlDB: sqlDB}
}

// DSN returns the shared MySQL DSN used by this repository.
func (r *Repository) DSN() string {
	return r.config.DSN()
}

// Migrations returns the versioned social schema ownership.
func (r *Repository) Migrations() []db.Migration {
	return []db.Migration{
		{
			ID: "001_social_core",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS social_friend_requests (
					request_id VARCHAR(64) PRIMARY KEY,
					from_player_id VARCHAR(64) NOT NULL,
					to_player_id VARCHAR(64) NOT NULL,
					status VARCHAR(16) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					UNIQUE KEY uq_social_friend_requests_pair (from_player_id, to_player_id),
					INDEX idx_social_friend_requests_to_player (to_player_id, status)
				);`,
				`CREATE TABLE IF NOT EXISTS social_friendships (
					player_id VARCHAR(64) NOT NULL,
					friend_player_id VARCHAR(64) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (player_id, friend_player_id),
					INDEX idx_social_friendships_friend (friend_player_id)
				);`,
				`CREATE TABLE IF NOT EXISTS social_blocks (
					player_id VARCHAR(64) NOT NULL,
					blocked_player_id VARCHAR(64) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (player_id, blocked_player_id),
					INDEX idx_social_blocks_blocked (blocked_player_id)
				);`,
			},
		},
		{
			ID: "002_social_relationship_metadata",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS social_friend_remarks (
					player_id VARCHAR(64) NOT NULL,
					friend_player_id VARCHAR(64) NOT NULL,
					remark VARCHAR(255) NOT NULL DEFAULT '',
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (player_id, friend_player_id),
					INDEX idx_social_friend_remarks_friend (friend_player_id)
				);`,
			},
		},
	}
}

// SchemaStatements returns the social-owned schema statements.
func (r *Repository) SchemaStatements() []string { return db.FlattenMigrations(r.Migrations()) }

// BootstrapSchema applies the social-owned schema against the configured MySQL connection.
func (r *Repository) BootstrapSchema(ctx context.Context) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	return db.ApplyMySQLMigrations(ctx, r.sqlDB, "social", r.Migrations())
}

func (r *Repository) ListFriendRequests() ([]domain.FriendRequest, error) {
	if r == nil || r.sqlDB == nil { return nil, errors.New("mysql repository is not configured") }
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT request_id, from_player_id, to_player_id, status, created_at FROM social_friend_requests`)
	if err != nil { return nil, err }
	defer rows.Close()
	requests := make([]domain.FriendRequest, 0)
	for rows.Next() {
		var request domain.FriendRequest
		if err := rows.Scan(&request.ID, &request.FromPlayerID, &request.ToPlayerID, &request.Status, &request.CreatedAt); err != nil { return nil, err }
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil { return nil, err }
	slices.SortFunc(requests, func(a, b domain.FriendRequest) int {
		if !a.CreatedAt.Equal(b.CreatedAt) { if a.CreatedAt.Before(b.CreatedAt) { return -1 }; return 1 }
		switch { case a.ID < b.ID: return -1; case a.ID > b.ID: return 1; default: return 0 }
	})
	return requests, nil
}

func (r *Repository) SaveFriendRequest(request domain.FriendRequest) error {
	if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO social_friend_requests (
			request_id, from_player_id, to_player_id, status, created_at
		) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			from_player_id = VALUES(from_player_id),
			to_player_id = VALUES(to_player_id),
			status = VALUES(status),
			created_at = VALUES(created_at)`,
		request.ID, request.FromPlayerID, request.ToPlayerID, request.Status, request.CreatedAt.UTC())
	return err
}

func (r *Repository) GetFriendRequest(requestID string) (domain.FriendRequest, bool, error) {
	if r == nil || r.sqlDB == nil { return domain.FriendRequest{}, false, errors.New("mysql repository is not configured") }
	row := r.sqlDB.QueryRowContext(context.Background(), `SELECT request_id, from_player_id, to_player_id, status, created_at FROM social_friend_requests WHERE request_id = ?`, requestID)
	var request domain.FriendRequest
	if err := row.Scan(&request.ID, &request.FromPlayerID, &request.ToPlayerID, &request.Status, &request.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return domain.FriendRequest{}, false, nil }
		return domain.FriendRequest{}, false, err
	}
	return request, true, nil
}

func (r *Repository) ListFriends(playerID string) ([]string, error) {
	if r == nil || r.sqlDB == nil { return nil, errors.New("mysql repository is not configured") }
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT friend_player_id FROM social_friendships WHERE player_id = ?`, playerID)
	if err != nil { return nil, err }
	defer rows.Close()
	friends := make([]string, 0)
	for rows.Next() {
		var friendID string
		if err := rows.Scan(&friendID); err != nil { return nil, err }
		friends = append(friends, friendID)
	}
	if err := rows.Err(); err != nil { return nil, err }
	slices.Sort(friends)
	return friends, nil
}

func (r *Repository) SaveFriendship(relationship domain.FriendRelationship) error {
	if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO social_friendships (player_id, friend_player_id) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE friend_player_id = VALUES(friend_player_id)`, relationship.PlayerID, relationship.FriendID)
	return err
}

func (r *Repository) ListBlocks(playerID string) ([]string, error) {
	if r == nil || r.sqlDB == nil { return nil, errors.New("mysql repository is not configured") }
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT blocked_player_id FROM social_blocks WHERE player_id = ?`, playerID)
	if err != nil { return nil, err }
	defer rows.Close()
	blocked := make([]string, 0)
	for rows.Next() {
		var blockedID string
		if err := rows.Scan(&blockedID); err != nil { return nil, err }
		blocked = append(blocked, blockedID)
	}
	if err := rows.Err(); err != nil { return nil, err }
	slices.Sort(blocked)
	return blocked, nil
}

func (r *Repository) SaveBlock(block domain.BlockRelationship) error {
	if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO social_blocks (player_id, blocked_player_id, created_at) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE created_at = VALUES(created_at)`, block.PlayerID, block.BlockedID, block.CreatedAt.UTC())
	return err
}

func (r *Repository) ListRemarks(playerID string) ([]domain.FriendRemark, error) {
	if r == nil || r.sqlDB == nil { return nil, errors.New("mysql repository is not configured") }
	rows, err := r.sqlDB.QueryContext(context.Background(), `SELECT player_id, friend_player_id, remark, updated_at FROM social_friend_remarks WHERE player_id = ?`, playerID)
	if err != nil { return nil, err }
	defer rows.Close()
	remarks := make([]domain.FriendRemark, 0)
	for rows.Next() {
		var remark domain.FriendRemark
		if err := rows.Scan(&remark.PlayerID, &remark.FriendID, &remark.Remark, &remark.UpdatedAt); err != nil { return nil, err }
		remarks = append(remarks, remark)
	}
	if err := rows.Err(); err != nil { return nil, err }
	slices.SortFunc(remarks, func(a, b domain.FriendRemark) int {
		switch { case a.FriendID < b.FriendID: return -1; case a.FriendID > b.FriendID: return 1; default: return 0 }
	})
	return remarks, nil
}

func (r *Repository) SaveRemark(remark domain.FriendRemark) error {
	if r == nil || r.sqlDB == nil { return errors.New("mysql repository is not configured") }
	_, err := r.sqlDB.ExecContext(context.Background(), `INSERT INTO social_friend_remarks (player_id, friend_player_id, remark, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE remark = VALUES(remark), updated_at = VALUES(updated_at)`, remark.PlayerID, remark.FriendID, remark.Remark, remark.UpdatedAt.UTC())
	return err
}

func (r *Repository) GetRemark(playerID string, friendID string) (domain.FriendRemark, bool, error) {
	if r == nil || r.sqlDB == nil { return domain.FriendRemark{}, false, errors.New("mysql repository is not configured") }
	row := r.sqlDB.QueryRowContext(context.Background(), `SELECT player_id, friend_player_id, remark, updated_at FROM social_friend_remarks WHERE player_id = ? AND friend_player_id = ?`, playerID, friendID)
	var remark domain.FriendRemark
	if err := row.Scan(&remark.PlayerID, &remark.FriendID, &remark.Remark, &remark.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return domain.FriendRemark{}, false, nil }
		return domain.FriendRemark{}, false, err
	}
	return remark, true, nil
}
