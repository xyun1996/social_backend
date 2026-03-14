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
	if bootstrapOnlyEnabled() {
		logger.Info("bootstrap-only mode completed")
		return
	}

	mux := handler.NewAuthHTTPHandler(authService).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func bootstrapOnlyEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("BOOTSTRAP_ONLY")), "true")
}

func buildAuthService() (*service.AuthService, func(), error) {
	options := service.Options{
		AccessTokenTTL:  durationFromEnv("IDENTITY_ACCESS_TOKEN_TTL", time.Hour),
		RefreshTokenTTL: durationFromEnv("IDENTITY_REFRESH_TOKEN_TTL", 7*24*time.Hour),
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("IDENTITY_STORE")), "mysql") {
		return service.NewAuthServiceWithOptions(nil, nil, options), func() {}, nil
	}

	mysqlConfig := db.LoadMySQLConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sqlDB, err := db.OpenMySQL(ctx, mysqlConfig)
	if err != nil {
		return nil, func() {}, err
	}
	repo := mysqlrepo.NewRepository(mysqlConfig, sqlDB)
	if strings.EqualFold(strings.TrimSpace(os.Getenv("IDENTITY_AUTO_MIGRATE")), "true") {
		if err := repo.BootstrapSchema(ctx); err != nil {
			_ = sqlDB.Close()
			return nil, func() {}, err
		}
	}
	return service.NewAuthServiceWithOptions(repo, repo, options), func() {
		_ = sqlDB.Close()
	}, nil
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}
