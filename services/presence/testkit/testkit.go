package testkit

import (
	"net/http/httptest"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/db"
	presencehandler "github.com/xyun1996/social_backend/services/presence/internal/handler"
	redisstore "github.com/xyun1996/social_backend/services/presence/internal/repo/redis"
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

// NewDurableServer constructs a presence HTTP server backed by Redis state.
func NewDurableServer(redisConfig db.RedisConfig, client *redis.Client) *Server {
	store := redisstore.NewStore(redisConfig, client)
	presence := presenceservice.NewPresenceServiceWithStore(store)
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
