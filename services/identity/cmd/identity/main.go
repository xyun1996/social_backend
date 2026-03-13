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
	"github.com/xyun1996/social_backend/services/identity/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/identity/internal/repo/mysql"
	"github.com/xyun1996/social_backend/services/identity/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("identity", ":8081")
	logger := logging.New(cfg.Name, cfg.Env)
	authService, cleanup, err := buildAuthService()
	if err != nil {
		logger.Error("failed to initialize identity service", "error", err)
		panic(err)
	}
	defer cleanup()

	mux := handler.NewAuthHTTPHandler(authService).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildAuthService() (*service.AuthService, func(), error) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("IDENTITY_STORE")), "mysql") {
		return service.NewAuthService(), func() {}, nil
	}

	mysqlConfig := db.LoadMySQLConfig()
	sqlDB, err := sql.Open("mysql", mysqlConfig.DSN())
	if err != nil {
		return nil, func() {}, err
	}

	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, func() {}, err
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("IDENTITY_AUTO_MIGRATE")), "true") {
		if err := repo.BootstrapSchema(ctx); err != nil {
			_ = sqlDB.Close()
			return nil, func() {}, err
		}
	}
	return service.NewAuthServiceWithStores(repo, repo), func() {
		_ = sqlDB.Close()
	}, nil
}
