package testkit

import (
	"net/http/httptest"

	chatclient "github.com/xyun1996/social_backend/services/gateway/internal/client/chat"
	identityclient "github.com/xyun1996/social_backend/services/gateway/internal/client/identity"
	presenceclient "github.com/xyun1996/social_backend/services/gateway/internal/client/presence"
	gatewayhandler "github.com/xyun1996/social_backend/services/gateway/internal/handler"
)

// Server hosts an in-memory gateway service for integration tests.
type Server struct {
	server *httptest.Server
}

// NewServer constructs an in-memory gateway HTTP server.
func NewServer(identityBaseURL string, presenceBaseURL string, chatBaseURL string) *Server {
	server := httptest.NewServer(gatewayhandler.NewHTTPHandler(
		identityclient.NewHTTPClient(identityBaseURL),
		presenceclient.NewHTTPClient(presenceBaseURL),
		chatclient.NewHTTPClient(chatBaseURL),
	).Routes())
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
