package mysql

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/social/internal/domain"
)

func TestRepositoryBootstrapSchema(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("social", "001_social_core").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.SchemaStatements() {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("social", "001_social_core").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.BootstrapSchema(context.Background()); err != nil {
		t.Fatalf("BootstrapSchema returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositoryFriendRequestLifecycle(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	createdAt := time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)
	request := domain.FriendRequest{
		ID:           "req-1",
		FromPlayerID: "p1",
		ToPlayerID:   "p2",
		Status:       "pending",
		CreatedAt:    createdAt,
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO social_friend_requests (
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
			created_at = VALUES(created_at)`)).
		WithArgs(request.ID, request.FromPlayerID, request.ToPlayerID, request.Status, request.CreatedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.SaveFriendRequest(request); err != nil {
		t.Fatalf("SaveFriendRequest returned error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"request_id", "from_player_id", "to_player_id", "status", "created_at"}).
		AddRow("req-2", "p3", "p2", "pending", createdAt.Add(time.Minute)).
		AddRow("req-1", "p1", "p2", "pending", createdAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT request_id, from_player_id, to_player_id, status, created_at
		 FROM social_friend_requests`)).
		WillReturnRows(rows)

	requests, err := repo.ListFriendRequests()
	if err != nil {
		t.Fatalf("ListFriendRequests returned error: %v", err)
	}
	if len(requests) != 2 || requests[0].ID != "req-1" || requests[1].ID != "req-2" {
		t.Fatalf("unexpected requests: %+v", requests)
	}

	row := sqlmock.NewRows([]string{"request_id", "from_player_id", "to_player_id", "status", "created_at"}).
		AddRow(request.ID, request.FromPlayerID, request.ToPlayerID, request.Status, request.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT request_id, from_player_id, to_player_id, status, created_at
		 FROM social_friend_requests
		 WHERE request_id = ?`)).
		WithArgs(request.ID).
		WillReturnRows(row)

	loaded, ok, err := repo.GetFriendRequest(request.ID)
	if err != nil {
		t.Fatalf("GetFriendRequest returned error: %v", err)
	}
	if !ok || loaded.ID != request.ID {
		t.Fatalf("unexpected loaded request: %+v (ok=%v)", loaded, ok)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositoryFriendshipsAndBlocks(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	blockedAt := time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO social_friendships (
			player_id,
			friend_player_id
		) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			friend_player_id = VALUES(friend_player_id)`)).
		WithArgs("p1", "p2").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.SaveFriendship(domain.FriendRelationship{PlayerID: "p1", FriendID: "p2"}); err != nil {
		t.Fatalf("SaveFriendship returned error: %v", err)
	}

	friendRows := sqlmock.NewRows([]string{"friend_player_id"}).
		AddRow("p3").
		AddRow("p2")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT friend_player_id
		 FROM social_friendships
		 WHERE player_id = ?`)).
		WithArgs("p1").
		WillReturnRows(friendRows)

	friends, err := repo.ListFriends("p1")
	if err != nil {
		t.Fatalf("ListFriends returned error: %v", err)
	}
	if len(friends) != 2 || friends[0] != "p2" || friends[1] != "p3" {
		t.Fatalf("unexpected friends: %+v", friends)
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO social_blocks (
			player_id,
			blocked_player_id,
			created_at
		) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			created_at = VALUES(created_at)`)).
		WithArgs("p1", "p9", blockedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.SaveBlock(domain.BlockRelationship{PlayerID: "p1", BlockedID: "p9", CreatedAt: blockedAt}); err != nil {
		t.Fatalf("SaveBlock returned error: %v", err)
	}

	blockRows := sqlmock.NewRows([]string{"blocked_player_id"}).
		AddRow("p9").
		AddRow("p4")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT blocked_player_id
		 FROM social_blocks
		 WHERE player_id = ?`)).
		WithArgs("p1").
		WillReturnRows(blockRows)

	blocks, err := repo.ListBlocks("p1")
	if err != nil {
		t.Fatalf("ListBlocks returned error: %v", err)
	}
	if len(blocks) != 2 || blocks[0] != "p4" || blocks[1] != "p9" {
		t.Fatalf("unexpected blocks: %+v", blocks)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
