package testkit

import (
	"net/http/httptest"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	chatclient "github.com/xyun1996/social_backend/services/gateway/internal/client/chat"
	identityclient "github.com/xyun1996/social_backend/services/gateway/internal/client/identity"
	presenceclient "github.com/xyun1996/social_backend/services/gateway/internal/client/presence"
	gatewayhandler "github.com/xyun1996/social_backend/services/gateway/internal/handler"
	redisrepo "github.com/xyun1996/social_backend/services/gateway/internal/repo/redis"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
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

// NewDurableServer constructs a gateway HTTP server backed by Redis session state.
func NewDurableServer(redisConfig db.RedisConfig, client *redis.Client, identityBaseURL string, presenceBaseURL string, chatBaseURL string) *Server {
	introspector := identityclient.NewHTTPClient(identityBaseURL)
	reporter := presenceclient.NewHTTPClient(presenceBaseURL)
	chat := chatclient.NewHTTPClient(chatBaseURL)
	repo := redisrepo.NewRepository(redisConfig, client)
	realtime := gatewayservice.NewRealtimeServiceWithStores(repo, repo, introspector, reporter)
	server := httptest.NewServer(gatewayhandler.NewHTTPHandlerWithRealtime(introspector, reporter, chat, realtime).Routes())
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
