package app

import (
	"net/http"

	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/social-core/internal/foundation/contracts"
	"github.com/xyun1996/social_backend/services/social-core/internal/modules"
)

// Runtime bundles the product-rebuild foundation seams for social-core.
type Runtime struct {
	Registry   modules.Registry
	Authorizer contracts.Authorizer
	Audit      contracts.AuditSink
	Tx         contracts.TxManager
	Jobs       contracts.JobEnqueuer
}

// NewRuntime creates the minimum runtime shape that future product-grade
// modules will register against.
func NewRuntime() Runtime {
	return Runtime{
		Registry: modules.NewRegistry(),
	}
}

// MountRuntimeEndpoints exposes the rebuild inventory so progress stays tied
// to the new runtime instead of the frozen prototype services.
func (r Runtime) MountRuntimeEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/runtime/status", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, map[string]any{
			"runtime": "social-core",
			"phase":   "product-rebuild",
			"modules": r.Registry.Descriptors(),
			"foundation": map[string]bool{
				"authorizer": r.Authorizer != nil,
				"audit_sink": r.Audit != nil,
				"tx_manager": r.Tx != nil,
				"job_queue":  r.Jobs != nil,
			},
		})
	})
}
