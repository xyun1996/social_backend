package testkit

import (
	"context"
	"fmt"
	"net/http/httptest"

	workerchatclient "github.com/xyun1996/social_backend/services/worker/internal/client/chat"
	workerinviteclient "github.com/xyun1996/social_backend/services/worker/internal/client/invite"
	workerhandler "github.com/xyun1996/social_backend/services/worker/internal/handler"
	workerjobs "github.com/xyun1996/social_backend/services/worker/internal/jobs"
	workerservice "github.com/xyun1996/social_backend/services/worker/internal/service"
)

// Server hosts an in-memory worker service for integration tests.
type Server struct {
	server *httptest.Server
	worker *workerservice.WorkerService
}

// ExecutionSummary is a public wrapper around worker execution counters.
type ExecutionSummary struct {
	Processed int
	Completed int
	Failed    int
}

// NewServer constructs an in-memory worker HTTP server.
func NewServer() *Server {
	worker := workerservice.NewWorkerService()
	server := httptest.NewServer(workerhandler.NewHTTPHandler(worker).Routes())
	return &Server{server: server, worker: worker}
}

// URL returns the HTTP base URL.
func (s *Server) URL() string {
	return s.server.URL
}

// Close shuts down the test server.
func (s *Server) Close() {
	s.server.Close()
}

// RegisterInviteExpireHandler wires the invite expiry job handler to an invite base URL.
func (s *Server) RegisterInviteExpireHandler(inviteBaseURL string) {
	s.worker.RegisterHandler("invite.expire", workerjobs.NewInviteExpireHandler(workerinviteclient.NewHTTPClient(inviteBaseURL)).Handle)
}

// RegisterChatOfflineDeliveryHandler wires the offline delivery job handler to a chat base URL.
func (s *Server) RegisterChatOfflineDeliveryHandler(chatBaseURL string) {
	s.worker.RegisterHandler("chat.offline_delivery", workerjobs.NewChatOfflineDeliveryHandler(workerchatclient.NewHTTPClient(chatBaseURL)).Handle)
}

// ExecuteUntilEmpty drains worker jobs and returns a public summary.
func (s *Server) ExecuteUntilEmpty(ctx context.Context, workerID string, jobType string, limit int) (ExecutionSummary, error) {
	result, appErr := s.worker.ExecuteUntilEmpty(ctx, workerID, jobType, limit)
	if appErr != nil {
		return ExecutionSummary{}, fmt.Errorf("%s: %s", appErr.Code, appErr.Message)
	}

	return ExecutionSummary{
		Processed: result.Processed,
		Completed: result.Completed,
		Failed:    result.Failed,
	}, nil
}
