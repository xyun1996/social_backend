package testkit

import (
	"database/sql"
	"net/http/httptest"

	"github.com/xyun1996/social_backend/pkg/db"
	guildclient "github.com/xyun1996/social_backend/services/ops/internal/client/guild"
	partyclient "github.com/xyun1996/social_backend/services/ops/internal/client/party"
	presenceclient "github.com/xyun1996/social_backend/services/ops/internal/client/presence"
	socialclient "github.com/xyun1996/social_backend/services/ops/internal/client/social"
	workerclient "github.com/xyun1996/social_backend/services/ops/internal/client/worker"
	opshandler "github.com/xyun1996/social_backend/services/ops/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/ops/internal/repo/mysql"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

// Server hosts an ops HTTP service for integration tests.
type Server struct {
	server *httptest.Server
}

// NewServer constructs an ops HTTP server backed by HTTP clients only.
func NewServer(presenceBaseURL string, partyBaseURL string, guildBaseURL string, workerBaseURL string, socialBaseURL string) *Server {
	ops := opsservice.NewOpsService(
		presenceclient.NewHTTPClient(presenceBaseURL),
		partyclient.NewHTTPClient(partyBaseURL),
		guildclient.NewHTTPClient(guildBaseURL),
		workerclient.NewHTTPClient(workerBaseURL),
		socialclient.NewHTTPClient(socialBaseURL),
		nil,
	)
	server := httptest.NewServer(opshandler.NewHTTPHandler(ops).Routes())
	return &Server{server: server}
}

// NewDurableServer constructs an ops HTTP server with optional MySQL bootstrap visibility.
func NewDurableServer(mysqlConfig db.MySQLConfig, sqlDB *sql.DB, presenceBaseURL string, partyBaseURL string, guildBaseURL string, workerBaseURL string, socialBaseURL string) *Server {
	ops := opsservice.NewOpsService(
		presenceclient.NewHTTPClient(presenceBaseURL),
		partyclient.NewHTTPClient(partyBaseURL),
		guildclient.NewHTTPClient(guildBaseURL),
		workerclient.NewHTTPClient(workerBaseURL),
		socialclient.NewHTTPClient(socialBaseURL),
		mysqlrepo.NewBootstrapReader(sqlDB),
	)
	_ = mysqlConfig
	server := httptest.NewServer(opshandler.NewHTTPHandler(ops).Routes())
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
