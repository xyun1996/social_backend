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
	"github.com/xyun1996/social_backend/services/social/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/social/internal/repo/mysql"
	"github.com/xyun1996/social_backend/services/social/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("social", ":8082")
	logger := logging.New(cfg.Name, cfg.Env)
	socialService, cleanup, err := buildSocialService()
	if err != nil {
		logger.Error("failed to initialize social service", "error", err)
		panic(err)
	}
	defer cleanup()
	if bootstrapOnlyEnabled() {
		logger.Info("bootstrap-only mode completed")
		return
	}

	mux := handler.NewHTTPHandler(socialService).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func bootstrapOnlyEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("BOOTSTRAP_ONLY")), "true")
}

func buildSocialService() (*service.SocialService, func(), error) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("SOCIAL_STORE")), "mysql") {
		return service.NewSocialService(), func() {}, nil
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
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SOCIAL_AUTO_MIGRATE")), "true") {
		if err := repo.BootstrapSchema(ctx); err != nil {
			_ = sqlDB.Close()
			return nil, func() {}, err
		}
	}

	return service.NewSocialServiceWithStores(repo, repo, repo), func() {
		_ = sqlDB.Close()
	}, nil
}
