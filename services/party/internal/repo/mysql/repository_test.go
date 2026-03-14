package mysql

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/services/party/internal/domain"
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
		WithArgs("party", "001_party_core").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.Migrations()[0].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("party", "001_party_core").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("party", "002_party_queue").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.Migrations()[1].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("party", "002_party_queue").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM schema_migrations WHERE service_name = ? AND migration_id = ? LIMIT 1")).
		WithArgs("party", "003_party_assignment").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	for _, statement := range repo.Migrations()[2].Statements {
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (service_name, migration_id) VALUES (?, ?)")).
		WithArgs("party", "003_party_assignment").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.BootstrapSchema(context.Background()); err != nil {
		t.Fatalf("BootstrapSchema returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySaveAndLoadParty(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	party := domain.Party{
		ID:        "party-1",
		LeaderID:  "p1",
		MemberIDs: []string{"p1", "p2"},
		CreatedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO party_parties (party_id, leader_id, created_at)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   leader_id = VALUES(leader_id),
		   created_at = VALUES(created_at)`)).
		WithArgs(party.ID, party.LeaderID, party.CreatedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM party_members WHERE party_id = ?`)).
		WithArgs(party.ID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO party_members (party_id, player_id) VALUES (?, ?)`)).
		WithArgs(party.ID, "p1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO party_members (party_id, player_id) VALUES (?, ?)`)).
		WithArgs(party.ID, "p2").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.SaveParty(party); err != nil {
		t.Fatalf("SaveParty returned error: %v", err)
	}

	row := sqlmock.NewRows([]string{"party_id", "leader_id", "created_at"}).
		AddRow(party.ID, party.LeaderID, party.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT party_id, leader_id, created_at FROM party_parties WHERE party_id = ?`)).
		WithArgs(party.ID).
		WillReturnRows(row)
	memberRows := sqlmock.NewRows([]string{"player_id"}).AddRow("p2").AddRow("p1")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT player_id FROM party_members WHERE party_id = ?`)).
		WithArgs(party.ID).
		WillReturnRows(memberRows)

	loaded, ok, err := repo.GetParty(party.ID)
	if err != nil {
		t.Fatalf("GetParty returned error: %v", err)
	}
	if !ok || len(loaded.MemberIDs) != 2 || loaded.MemberIDs[0] != "p1" {
		t.Fatalf("unexpected loaded party: %+v", loaded)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySaveAndListReadyStates(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	state := domain.ReadyState{
		PartyID:   "party-1",
		PlayerID:  "p1",
		IsReady:   true,
		UpdatedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO party_ready_states (party_id, player_id, is_ready, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   is_ready = VALUES(is_ready),
		   updated_at = VALUES(updated_at)`)).
		WithArgs(state.PartyID, state.PlayerID, state.IsReady, state.UpdatedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.SaveReadyState(state); err != nil {
		t.Fatalf("SaveReadyState returned error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"party_id", "player_id", "is_ready", "updated_at"}).
		AddRow("party-1", "p2", false, state.UpdatedAt.Add(time.Minute)).
		AddRow("party-1", "p1", true, state.UpdatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT party_id, player_id, is_ready, updated_at
		 FROM party_ready_states
		 WHERE party_id = ?`)).
		WithArgs(state.PartyID).
		WillReturnRows(rows)

	states, err := repo.ListReadyStates(state.PartyID)
	if err != nil {
		t.Fatalf("ListReadyStates returned error: %v", err)
	}
	if len(states) != 2 || states[0].PlayerID != "p1" || !states[0].IsReady {
		t.Fatalf("unexpected ready states: %+v", states)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositoryDeleteReadyState(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM party_ready_states WHERE party_id = ? AND player_id = ?`)).
		WithArgs("party-1", "p2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteReadyState("party-1", "p2"); err != nil {
		t.Fatalf("DeleteReadyState returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySaveAndDeleteQueueState(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	state := domain.QueueState{
		PartyID:   "party-1",
		QueueName: "casual-2v2",
		Status:    "queued",
		JoinedBy:  "p1",
		JoinedAt:  time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO party_queue_states (party_id, queue_name, status, joined_by, joined_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   queue_name = VALUES(queue_name),
		   status = VALUES(status),
		   joined_by = VALUES(joined_by),
		   joined_at = VALUES(joined_at)`)).
		WithArgs(state.PartyID, state.QueueName, state.Status, state.JoinedBy, state.JoinedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.SaveQueueState(state); err != nil {
		t.Fatalf("SaveQueueState returned error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"party_id", "queue_name", "status", "joined_by", "joined_at"}).
		AddRow(state.PartyID, state.QueueName, state.Status, state.JoinedBy, state.JoinedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT party_id, queue_name, status, joined_by, joined_at
		 FROM party_queue_states
		 WHERE party_id = ?`)).
		WithArgs(state.PartyID).
		WillReturnRows(rows)

	loaded, ok, err := repo.GetQueueState(state.PartyID)
	if err != nil {
		t.Fatalf("GetQueueState returned error: %v", err)
	}
	if !ok || loaded.QueueName != state.QueueName || loaded.JoinedBy != state.JoinedBy {
		t.Fatalf("unexpected loaded queue state: %+v", loaded)
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM party_queue_states WHERE party_id = ?`)).
		WithArgs(state.PartyID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteQueueState(state.PartyID); err != nil {
		t.Fatalf("DeleteQueueState returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySaveAndDeleteQueueAssignment(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer sqlDB.Close()

	repo := NewRepository(db.MySQLConfig{}, sqlDB)
	assignment := domain.QueueAssignment{
		TicketID:       "ticket:party-1:casual-2v2:1760000000",
		PartyID:        "party-1",
		QueueName:      "casual-2v2",
		MatchID:        "match-1",
		Status:         "assigned",
		ServerID:       "game-1",
		ConnectionHint: "wss://game-1/session/match-1",
		AssignedAt:     time.Date(2026, 3, 13, 12, 5, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO party_queue_assignments (party_id, ticket_id, queue_name, match_id, status, server_id, connection_hint, assigned_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   ticket_id = VALUES(ticket_id),
		   queue_name = VALUES(queue_name),
		   match_id = VALUES(match_id),
		   status = VALUES(status),
		   server_id = VALUES(server_id),
		   connection_hint = VALUES(connection_hint),
		   assigned_at = VALUES(assigned_at)`)).
		WithArgs(
			assignment.PartyID,
			assignment.TicketID,
			assignment.QueueName,
			assignment.MatchID,
			assignment.Status,
			assignment.ServerID,
			assignment.ConnectionHint,
			assignment.AssignedAt.UTC(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.SaveQueueAssignment(assignment); err != nil {
		t.Fatalf("SaveQueueAssignment returned error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"party_id", "ticket_id", "queue_name", "match_id", "status", "server_id", "connection_hint", "assigned_at"}).
		AddRow(
			assignment.PartyID,
			assignment.TicketID,
			assignment.QueueName,
			assignment.MatchID,
			assignment.Status,
			assignment.ServerID,
			assignment.ConnectionHint,
			assignment.AssignedAt,
		)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT party_id, ticket_id, queue_name, match_id, status, server_id, connection_hint, assigned_at
		 FROM party_queue_assignments
		 WHERE party_id = ?`)).
		WithArgs(assignment.PartyID).
		WillReturnRows(rows)

	loaded, ok, err := repo.GetQueueAssignment(assignment.PartyID)
	if err != nil {
		t.Fatalf("GetQueueAssignment returned error: %v", err)
	}
	if !ok || loaded.MatchID != assignment.MatchID || loaded.ServerID != assignment.ServerID {
		t.Fatalf("unexpected loaded queue assignment: %+v", loaded)
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM party_queue_assignments WHERE party_id = ?`)).
		WithArgs(assignment.PartyID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteQueueAssignment(assignment.PartyID); err != nil {
		t.Fatalf("DeleteQueueAssignment returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
