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

// SchemaStatements returns the social-owned schema statements.
func (r *Repository) SchemaStatements() []string {
	return []string{
		`CREATE TABLE social_friend_requests (
			request_id VARCHAR(64) PRIMARY KEY,
			from_player_id VARCHAR(64) NOT NULL,
			to_player_id VARCHAR(64) NOT NULL,
			status VARCHAR(16) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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

// BootstrapSchema applies the social-owned schema against the configured MySQL connection.
func (r *Repository) BootstrapSchema(ctx context.Context) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}
	return applySchema(ctx, r.sqlDB, r.SchemaStatements())
}

// ListFriendRequests returns all persisted requests ordered by created time then id.
func (r *Repository) ListFriendRequests() ([]domain.FriendRequest, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT request_id, from_player_id, to_player_id, status, created_at
		 FROM social_friend_requests`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]domain.FriendRequest, 0)
	for rows.Next() {
		var request domain.FriendRequest
		if err := rows.Scan(
			&request.ID,
			&request.FromPlayerID,
			&request.ToPlayerID,
			&request.Status,
			&request.CreatedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(requests, func(a domain.FriendRequest, b domain.FriendRequest) int {
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
	return requests, nil
}

// SaveFriendRequest upserts a friend request lifecycle row.
func (r *Repository) SaveFriendRequest(request domain.FriendRequest) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO social_friend_requests (
			request_id,
			from_player_id,
			to_player_id,
			status,
			created_at
		) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			from_player_id = VALUES(from_player_id),
			to_player_id = VALUES(to_player_id),
			status = VALUES(status),
			created_at = VALUES(created_at)`,
		request.ID,
		request.FromPlayerID,
		request.ToPlayerID,
		request.Status,
		request.CreatedAt.UTC(),
	)
	return err
}

// GetFriendRequest loads a request by id.
func (r *Repository) GetFriendRequest(requestID string) (domain.FriendRequest, bool, error) {
	if r == nil || r.sqlDB == nil {
		return domain.FriendRequest{}, false, errors.New("mysql repository is not configured")
	}

	row := r.sqlDB.QueryRowContext(
		context.Background(),
		`SELECT request_id, from_player_id, to_player_id, status, created_at
		 FROM social_friend_requests
		 WHERE request_id = ?`,
		requestID,
	)

	var request domain.FriendRequest
	if err := row.Scan(
		&request.ID,
		&request.FromPlayerID,
		&request.ToPlayerID,
		&request.Status,
		&request.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FriendRequest{}, false, nil
		}
		return domain.FriendRequest{}, false, err
	}
	return request, true, nil
}

// ListFriends returns a stable friend list for the given player.
func (r *Repository) ListFriends(playerID string) ([]string, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT friend_player_id
		 FROM social_friendships
		 WHERE player_id = ?`,
		playerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := make([]string, 0)
	for rows.Next() {
		var friendID string
		if err := rows.Scan(&friendID); err != nil {
			return nil, err
		}
		friends = append(friends, friendID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.Sort(friends)
	return friends, nil
}

// SaveFriendship upserts an accepted friendship row.
func (r *Repository) SaveFriendship(relationship domain.FriendRelationship) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO social_friendships (
			player_id,
			friend_player_id
		) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			friend_player_id = VALUES(friend_player_id)`,
		relationship.PlayerID,
		relationship.FriendID,
	)
	return err
}

// ListBlocks returns a stable block list for the given player.
func (r *Repository) ListBlocks(playerID string) ([]string, error) {
	if r == nil || r.sqlDB == nil {
		return nil, errors.New("mysql repository is not configured")
	}

	rows, err := r.sqlDB.QueryContext(
		context.Background(),
		`SELECT blocked_player_id
		 FROM social_blocks
		 WHERE player_id = ?`,
		playerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocked := make([]string, 0)
	for rows.Next() {
		var blockedID string
		if err := rows.Scan(&blockedID); err != nil {
			return nil, err
		}
		blocked = append(blocked, blockedID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.Sort(blocked)
	return blocked, nil
}

// SaveBlock upserts a block relationship row.
func (r *Repository) SaveBlock(block domain.BlockRelationship) error {
	if r == nil || r.sqlDB == nil {
		return errors.New("mysql repository is not configured")
	}

	_, err := r.sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO social_blocks (
			player_id,
			blocked_player_id,
			created_at
		) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			created_at = VALUES(created_at)`,
		block.PlayerID,
		block.BlockedID,
		block.CreatedAt.UTC(),
	)
	return err
}

func applySchema(ctx context.Context, exec schemaExecutor, statements []string) error {
	for _, statement := range statements {
		if _, err := exec.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
