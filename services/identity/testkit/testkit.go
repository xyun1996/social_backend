package testkit

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	identityhandler "github.com/xyun1996/social_backend/services/identity/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/identity/internal/repo/mysql"
	identityservice "github.com/xyun1996/social_backend/services/identity/internal/service"
)

// Server hosts an in-memory identity service for integration tests.
type Server struct {
	server *httptest.Server
}

// NewServer constructs an in-memory identity HTTP server.
func NewServer() *Server {
	auth := identityservice.NewAuthService()
	server := httptest.NewServer(identityhandler.NewAuthHTTPHandler(auth).Routes())
	return &Server{server: server}
}

// NewDurableServer constructs an identity HTTP server backed by MySQL state.
func NewDurableServer(mysqlConfig db.MySQLConfig, sqlDB *sql.DB) *Server {
	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := repo.BootstrapSchema(ctx); err != nil {
		panic(err)
	}

	auth := identityservice.NewAuthServiceWithStores(repo, repo)
	server := httptest.NewServer(identityhandler.NewAuthHTTPHandler(auth).Routes())
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
