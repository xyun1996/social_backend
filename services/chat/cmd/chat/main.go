package main

import (
	"context"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/pkg/logging"
	guildclient "github.com/xyun1996/social_backend/services/chat/internal/client/guild"
	partyclient "github.com/xyun1996/social_backend/services/chat/internal/client/party"
	presenceclient "github.com/xyun1996/social_backend/services/chat/internal/client/presence"
	workerclient "github.com/xyun1996/social_backend/services/chat/internal/client/worker"
	"github.com/xyun1996/social_backend/services/chat/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/chat/internal/repo/mysql"
	"github.com/xyun1996/social_backend/services/chat/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("chat", ":8084")
	logger := logging.New(cfg.Name, cfg.Env)
	chatService, cleanup, err := buildChatService()
	if err != nil {
		logger.Error("failed to initialize chat service", "error", err)
		panic(err)
	}
	defer cleanup()
	if bootstrapOnlyEnabled() {
		logger.Info("bootstrap-only mode completed")
		return
	}

	mux := handler.NewHTTPHandler(chatService).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildChatService() (*service.ChatService, func(), error) {
	presenceBaseURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://127.0.0.1:8087")
	guildBaseURL := valueOrDefault(os.Getenv("GUILD_BASE_URL"), "http://127.0.0.1:8089")
	partyBaseURL := valueOrDefault(os.Getenv("PARTY_BASE_URL"), "http://127.0.0.1:8090")
	workerBaseURL := os.Getenv("WORKER_BASE_URL")

	var scheduler service.JobScheduler
	if workerBaseURL != "" {
		scheduler = workerclient.NewHTTPClient(workerBaseURL)
	}
	presence := presenceclient.NewHTTPClient(presenceBaseURL)
	guilds := guildclient.NewHTTPClient(guildBaseURL)
	parties := partyclient.NewHTTPClient(partyBaseURL)

	if !strings.EqualFold(strings.TrimSpace(os.Getenv("CHAT_STORE")), "mysql") {
		chatService := service.NewChatService(presence, scheduler)
		chatService.SetMembershipReaders(guilds, parties)
		return chatService, func() {}, nil
	}

	mysqlConfig := db.LoadMySQLConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sqlDB, err := db.OpenMySQL(ctx, mysqlConfig)
	if err != nil {
		return nil, func() {}, err
	}

	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	if strings.EqualFold(strings.TrimSpace(os.Getenv("CHAT_AUTO_MIGRATE")), "true") {
		if err := repo.BootstrapSchema(ctx); err != nil {
			_ = sqlDB.Close()
			return nil, func() {}, err
		}
	}

	chatService := service.NewChatServiceWithStores(repo, repo, repo, presence, scheduler)
	chatService.SetMembershipReaders(guilds, parties)
	return chatService, func() {
		_ = sqlDB.Close()
	}, nil
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func bootstrapOnlyEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("BOOTSTRAP_ONLY")), "true")
}
