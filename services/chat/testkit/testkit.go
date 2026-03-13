package testkit

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"time"

	"github.com/xyun1996/social_backend/pkg/db"
	presenceclient "github.com/xyun1996/social_backend/services/chat/internal/client/presence"
	workerclient "github.com/xyun1996/social_backend/services/chat/internal/client/worker"
	chathandler "github.com/xyun1996/social_backend/services/chat/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/chat/internal/repo/mysql"
	chatservice "github.com/xyun1996/social_backend/services/chat/internal/service"
)

// Server hosts an in-memory chat service for integration tests.
type Server struct {
	server *httptest.Server
	chat   *chatservice.ChatService
}

// NewServer constructs an in-memory chat HTTP server.
func NewServer(presenceBaseURL string, workerBaseURL string) *Server {
	var presence chatservice.PresenceReader
	if presenceBaseURL != "" {
		presence = presenceclient.NewHTTPClient(presenceBaseURL)
	}

	var scheduler chatservice.JobScheduler
	if workerBaseURL != "" {
		scheduler = workerclient.NewHTTPClient(workerBaseURL)
	}

	chat := chatservice.NewChatService(presence, scheduler)
	server := httptest.NewServer(chathandler.NewHTTPHandler(chat).Routes())
	return &Server{server: server, chat: chat}
}

// NewDurableServer constructs a chat HTTP server backed by MySQL stores.
func NewDurableServer(mysqlConfig db.MySQLConfig, sqlDB *sql.DB, presenceBaseURL string, workerBaseURL string) *Server {
	var presence chatservice.PresenceReader
	if presenceBaseURL != "" {
		presence = presenceclient.NewHTTPClient(presenceBaseURL)
	}

	var scheduler chatservice.JobScheduler
	if workerBaseURL != "" {
		scheduler = workerclient.NewHTTPClient(workerBaseURL)
	}

	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := repo.BootstrapSchema(ctx); err != nil {
		panic(err)
	}
	chat := chatservice.NewChatServiceWithStores(repo, repo, repo, presence, scheduler)
	server := httptest.NewServer(chathandler.NewHTTPHandler(chat).Routes())
	return &Server{server: server, chat: chat}
}

// URL returns the HTTP base URL.
func (s *Server) URL() string {
	return s.server.URL
}

// Close shuts down the test server.
func (s *Server) Close() {
	s.server.Close()
}

// OfflineDeliveryCount returns recorded offline delivery receipts for a conversation.
func (s *Server) OfflineDeliveryCount(conversationID string) int {
	return len(s.chat.ListOfflineDeliveries(conversationID))
}
