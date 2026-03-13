package testkit

import (
	"net/http/httptest"

	identityhandler "github.com/xyun1996/social_backend/services/identity/internal/handler"
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

// URL returns the HTTP base URL.
func (s *Server) URL() string {
	return s.server.URL
}

// Close shuts down the test server.
func (s *Server) Close() {
	s.server.Close()
}
