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
	chatclient "github.com/xyun1996/social_backend/services/guild/internal/client/chat"
	inviteclient "github.com/xyun1996/social_backend/services/guild/internal/client/invite"
	presenceclient "github.com/xyun1996/social_backend/services/guild/internal/client/presence"
	workerclient "github.com/xyun1996/social_backend/services/guild/internal/client/worker"
	"github.com/xyun1996/social_backend/services/guild/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/guild/internal/repo/mysql"
	"github.com/xyun1996/social_backend/services/guild/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("guild", ":8086")
	logger := logging.New(cfg.Name, cfg.Env)
	guilds, cleanup, err := buildGuildService()
	if err != nil {
		logger.Error("failed to initialize guild service", "error", err)
		panic(err)
	}
	defer cleanup()
	if bootstrapOnlyEnabled() {
		logger.Info("bootstrap-only mode completed")
		return
	}

	mux := handler.NewHTTPHandler(guilds).Routes()
	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildGuildService() (*service.GuildService, func(), error) {
	inviteURL := valueOrDefault(os.Getenv("INVITE_SERVICE_URL"), "http://localhost:8083")
	presenceURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://localhost:8087")
	chatURL := valueOrDefault(os.Getenv("CHAT_BASE_URL"), "http://localhost:8084")
	workerURL := strings.TrimSpace(os.Getenv("WORKER_BASE_URL"))
	invites := inviteclient.NewHTTPClient(inviteURL)
	presence := presenceclient.NewHTTPClient(presenceURL)
	chat := chatclient.NewHTTPClient(chatURL)

	var scheduler service.JobScheduler
	if workerURL != "" {
		scheduler = workerclient.NewHTTPClient(workerURL)
	}

	if !strings.EqualFold(strings.TrimSpace(os.Getenv("GUILD_STORE")), "mysql") {
		guilds := service.NewGuildService(invites, presence)
		guilds.SetRuntimeIntegrations(chat, scheduler)
		return guilds, func() {}, nil
	}

	mysqlConfig := db.LoadMySQLConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sqlDB, err := db.OpenMySQL(ctx, mysqlConfig)
	if err != nil {
		return nil, func() {}, err
	}

	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	if strings.EqualFold(strings.TrimSpace(os.Getenv("GUILD_AUTO_MIGRATE")), "true") {
		if err := repo.BootstrapSchema(ctx); err != nil {
			_ = sqlDB.Close()
			return nil, func() {}, err
		}
	}

	guilds := service.NewGuildServiceWithStore(repo, invites, presence)
	guilds.SetRuntimeIntegrations(chat, scheduler)
	return guilds, func() {
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
