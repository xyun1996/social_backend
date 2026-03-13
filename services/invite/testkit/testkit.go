package testkit

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	workerclient "github.com/xyun1996/social_backend/services/invite/internal/client/worker"
	invitehandler "github.com/xyun1996/social_backend/services/invite/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/invite/internal/repo/mysql"
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

// NewDurableServer constructs an invite HTTP server backed by MySQL state.
func NewDurableServer(mysqlConfig db.MySQLConfig, sqlDB *sql.DB, workerBaseURL string) *Server {
	var scheduler inviteservice.JobScheduler
	if workerBaseURL != "" {
		scheduler = workerclient.NewHTTPClient(workerBaseURL)
	}

	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := repo.BootstrapSchema(ctx); err != nil {
		panic(err)
	}
	invites := inviteservice.NewInviteServiceWithStore(repo, scheduler)
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
