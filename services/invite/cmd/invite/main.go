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
	workerclient "github.com/xyun1996/social_backend/services/invite/internal/client/worker"
	"github.com/xyun1996/social_backend/services/invite/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/invite/internal/repo/mysql"
	"github.com/xyun1996/social_backend/services/invite/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("invite", ":8083")
	logger := logging.New(cfg.Name, cfg.Env)
	inviteService, cleanup, err := buildInviteService()
	if err != nil {
		logger.Error("failed to initialize invite service", "error", err)
		panic(err)
	}
	defer cleanup()

	mux := handler.NewHTTPHandler(inviteService).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildInviteService() (*service.InviteService, func(), error) {
	workerURL := os.Getenv("WORKER_BASE_URL")

	var scheduler service.JobScheduler
	if workerURL != "" {
		scheduler = workerclient.NewHTTPClient(workerURL)
	}

	if !strings.EqualFold(strings.TrimSpace(os.Getenv("INVITE_STORE")), "mysql") {
		return service.NewInviteService(scheduler), func() {}, nil
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
	if strings.EqualFold(strings.TrimSpace(os.Getenv("INVITE_AUTO_MIGRATE")), "true") {
		if err := repo.BootstrapSchema(ctx); err != nil {
			_ = sqlDB.Close()
			return nil, func() {}, err
		}
	}

	return service.NewInviteServiceWithStore(repo, scheduler), func() {
		_ = sqlDB.Close()
	}, nil
}
