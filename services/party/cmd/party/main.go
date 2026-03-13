package main

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/pkg/logging"
	inviteclient "github.com/xyun1996/social_backend/services/party/internal/client/invite"
	presenceclient "github.com/xyun1996/social_backend/services/party/internal/client/presence"
	"github.com/xyun1996/social_backend/services/party/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/party/internal/repo/mysql"
	"github.com/xyun1996/social_backend/services/party/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("party", ":8085")
	logger := logging.New(cfg.Name, cfg.Env)
	parties, cleanup, err := buildPartyService()
	if err != nil {
		logger.Error("failed to initialize party service", "error", err)
		panic(err)
	}
	defer cleanup()
	if bootstrapOnlyEnabled() {
		logger.Info("bootstrap-only mode completed")
		return
	}

	mux := handler.NewHTTPHandler(parties).Routes()
	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildPartyService() (*service.PartyService, func(), error) {
	inviteURL := valueOrDefault(os.Getenv("INVITE_SERVICE_URL"), "http://localhost:8083")
	presenceURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://localhost:8087")
	invites := inviteclient.NewHTTPClient(inviteURL)
	presence := presenceclient.NewHTTPClient(presenceURL)

	if !strings.EqualFold(strings.TrimSpace(os.Getenv("PARTY_STORE")), "mysql") {
		return service.NewPartyService(invites, presence), func() {}, nil
	}

	mysqlConfig := db.LoadMySQLConfig()
	sqlDB, err := sql.Open("mysql", mysqlConfig.DSN())
	if err != nil {
		return nil, func() {}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, func() {}, err
	}

	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	if strings.EqualFold(strings.TrimSpace(os.Getenv("PARTY_AUTO_MIGRATE")), "true") {
		if err := repo.BootstrapSchema(ctx); err != nil {
			_ = sqlDB.Close()
			return nil, func() {}, err
		}
	}

	return service.NewPartyServiceWithStores(repo, repo, invites, presence), func() {
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
