package testkit

import (
	"net/http/httptest"

	workerclient "github.com/xyun1996/social_backend/services/invite/internal/client/worker"
	invitehandler "github.com/xyun1996/social_backend/services/invite/internal/handler"
	inviteservice "github.com/xyun1996/social_backend/services/invite/internal/service"
)

// Server hosts an in-memory invite service for integration tests.
type Server struct {
	server *httptest.Server
}

// NewServer constructs an in-memory invite HTTP server.
func NewServer(workerBaseURL string) *Server {
	var scheduler inviteservice.JobScheduler
	if workerBaseURL != "" {
		scheduler = workerclient.NewHTTPClient(workerBaseURL)
	}

	invites := inviteservice.NewInviteService(scheduler)
	server := httptest.NewServer(invitehandler.NewHTTPHandler(invites).Routes())
	return &Server{server: server}
}

// URL returns the HTTP base URL.
func (s *Server) URL() string {
	return s.server.URL
}

// Close shuts down the test server.
func (s *Server) Close() {
	s.server.Close()
}
