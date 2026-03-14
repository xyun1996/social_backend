package contracts

import "context"

// Principal represents the authenticated caller context inside the rebuilt
// product runtime.
type Principal struct {
	AccountID string
	PlayerID  string
	Roles     []string
}

// Authorizer centralizes runtime authorization decisions.
type Authorizer interface {
	Authorize(ctx context.Context, principal Principal, action string, resource string) error
}

// AuditSink records security-sensitive and business-critical actions.
type AuditSink interface {
	Record(ctx context.Context, event AuditEvent) error
}

// AuditEvent is the normalized shape expected by product-grade audit logging.
type AuditEvent struct {
	ActorAccountID string
	ActorPlayerID  string
	Action         string
	Resource       string
	Result         string
}

// TxManager describes the persistence boundary expected by product modules.
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
}

// JobEnqueuer captures async follow-up work without coupling modules to a
// specific worker transport.
type JobEnqueuer interface {
	Enqueue(ctx context.Context, job JobRequest) error
}

// JobRequest is the minimum cross-module async payload shape.
type JobRequest struct {
	Kind    string
	Key     string
	Payload map[string]any
}
