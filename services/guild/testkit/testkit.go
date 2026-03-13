package testkit

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	inviteclient "github.com/xyun1996/social_backend/services/guild/internal/client/invite"
	presenceclient "github.com/xyun1996/social_backend/services/guild/internal/client/presence"
	guildhandler "github.com/xyun1996/social_backend/services/guild/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/guild/internal/repo/mysql"
	guildservice "github.com/xyun1996/social_backend/services/guild/internal/service"
)

// Server hosts a guild HTTP service for integration tests.
type Server struct {
	server *httptest.Server
}

// NewServer constructs an in-memory guild HTTP server.
func NewServer(inviteBaseURL string, presenceBaseURL string) *Server {
	guilds := guildservice.NewGuildService(
		inviteclient.NewHTTPClient(inviteBaseURL),
		presenceclient.NewHTTPClient(presenceBaseURL),
	)
	server := httptest.NewServer(guildhandler.NewHTTPHandler(guilds).Routes())
	return &Server{server: server}
}

// NewDurableServer constructs a guild HTTP server backed by MySQL state.
func NewDurableServer(mysqlConfig db.MySQLConfig, sqlDB *sql.DB, inviteBaseURL string, presenceBaseURL string) *Server {
	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := repo.BootstrapSchema(ctx); err != nil {
		panic(err)
	}

	guilds := guildservice.NewGuildServiceWithStore(
		repo,
		inviteclient.NewHTTPClient(inviteBaseURL),
		presenceclient.NewHTTPClient(presenceBaseURL),
	)
	server := httptest.NewServer(guildhandler.NewHTTPHandler(guilds).Routes())
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
