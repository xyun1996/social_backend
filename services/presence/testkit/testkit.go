package testkit

import (
	"net/http/httptest"

	presencehandler "github.com/xyun1996/social_backend/services/presence/internal/handler"
	presenceservice "github.com/xyun1996/social_backend/services/presence/internal/service"
)

// Server hosts an in-memory presence service for integration tests.
type Server struct {
	server *httptest.Server
}

// NewServer constructs an in-memory presence HTTP server.
func NewServer() *Server {
	presence := presenceservice.NewPresenceService()
	server := httptest.NewServer(presencehandler.NewHTTPHandler(presence).Routes())
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
