package mysql

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type fakeSchemaExecutor struct {
	statements []string
	errAt      int
}

func (f *fakeSchemaExecutor) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	f.statements = append(f.statements, query)
	if f.errAt > 0 && len(f.statements) == f.errAt {
		return nil, errors.New("boom")
	}
	return nil, nil
}

func TestApplySchemaRunsAllStatementsInOrder(t *testing.T) {
	t.Parallel()

	exec := &fakeSchemaExecutor{}
	statements := []string{"stmt-1", "stmt-2"}
	if err := applySchema(context.Background(), exec, statements); err != nil {
		t.Fatalf("apply schema returned error: %v", err)
	}
	if len(exec.statements) != len(statements) || exec.statements[0] != "stmt-1" || exec.statements[1] != "stmt-2" {
		t.Fatalf("unexpected executed statements: %+v", exec.statements)
	}
}

func TestApplySchemaStopsOnStatementError(t *testing.T) {
	t.Parallel()

	exec := &fakeSchemaExecutor{errAt: 2}
	err := applySchema(context.Background(), exec, []string{"stmt-1", "stmt-2", "stmt-3"})
	if err == nil {
		t.Fatalf("expected apply schema to fail")
	}
	if len(exec.statements) != 2 {
		t.Fatalf("expected execution to stop at failing statement, got %+v", exec.statements)
	}
}
