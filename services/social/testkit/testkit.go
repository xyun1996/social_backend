package testkit

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	socialhandler "github.com/xyun1996/social_backend/services/social/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/social/internal/repo/mysql"
	socialservice "github.com/xyun1996/social_backend/services/social/internal/service"
)

// Server hosts a social HTTP service for integration tests.
type Server struct {
	server *httptest.Server
}

// NewServer constructs an in-memory social HTTP server.
func NewServer() *Server {
	social := socialservice.NewSocialService()
	server := httptest.NewServer(socialhandler.NewHTTPHandler(social).Routes())
	return &Server{server: server}
}

// NewDurableServer constructs a social HTTP server backed by MySQL state.
func NewDurableServer(mysqlConfig db.MySQLConfig, sqlDB *sql.DB) *Server {
	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := repo.BootstrapSchema(ctx); err != nil {
		panic(err)
	}
	social := socialservice.NewSocialServiceWithStores(repo, repo, repo, repo)
	server := httptest.NewServer(socialhandler.NewHTTPHandler(social).Routes())
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

