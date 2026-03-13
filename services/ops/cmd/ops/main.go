package main

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/pkg/logging"
	guildclient "github.com/xyun1996/social_backend/services/ops/internal/client/guild"
	partyclient "github.com/xyun1996/social_backend/services/ops/internal/client/party"
	presenceclient "github.com/xyun1996/social_backend/services/ops/internal/client/presence"
	socialclient "github.com/xyun1996/social_backend/services/ops/internal/client/social"
	workerclient "github.com/xyun1996/social_backend/services/ops/internal/client/worker"
	"github.com/xyun1996/social_backend/services/ops/internal/handler"
	mysqlrepo "github.com/xyun1996/social_backend/services/ops/internal/repo/mysql"
	redisrepo "github.com/xyun1996/social_backend/services/ops/internal/repo/redis"
	"github.com/xyun1996/social_backend/services/ops/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("ops", ":8088")
	logger := logging.New(cfg.Name, cfg.Env)
	presenceURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://localhost:8087")
	partyURL := valueOrDefault(os.Getenv("PARTY_BASE_URL"), "http://localhost:8085")
	guildURL := valueOrDefault(os.Getenv("GUILD_BASE_URL"), "http://localhost:8086")
	workerURL := valueOrDefault(os.Getenv("WORKER_BASE_URL"), "http://localhost:8089")
	socialURL := valueOrDefault(os.Getenv("SOCIAL_BASE_URL"), "http://localhost:8082")
	opsService, cleanup, err := buildOpsService(presenceURL, partyURL, guildURL, workerURL, socialURL)
	if err != nil {
		logger.Error("failed to initialize ops service", "error", err)
		panic(err)
	}
	defer cleanup()

	mux := handler.NewHTTPHandler(opsService).Routes()

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildOpsService(presenceURL string, partyURL string, guildURL string, workerURL string, socialURL string) (*service.OpsService, func(), error) {
	var bootstrapReader service.BootstrapReader
	var redisRuntimeReader service.RedisRuntimeReader
	cleanups := make([]func(), 0, 2)

	if strings.EqualFold(strings.TrimSpace(os.Getenv("OPS_MYSQL_STATUS")), "true") {
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

		bootstrapReader = mysqlrepo.NewBootstrapReader(sqlDB)
		cleanups = append(cleanups, func() {
			_ = sqlDB.Close()
		})
	}

	if strings.EqualFold(strings.TrimSpace(os.Getenv("OPS_REDIS_STATUS")), "true") {
		redisConfig := db.LoadRedisConfig()
		redisClient := redis.NewClient(&redis.Options{
			Addr:     redisConfig.Addr,
			Username: redisConfig.Username,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			_ = redisClient.Close()
			return nil, func() {}, err
		}

		redisRuntimeReader = redisrepo.NewRuntimeReader(redisConfig, redisClient)
		cleanups = append(cleanups, func() {
			_ = redisClient.Close()
		})
	}

	cleanup := func() {
		for _, closer := range cleanups {
			closer()
		}
	}

	return service.NewOpsService(
		presenceclient.NewHTTPClient(presenceURL),
		partyclient.NewHTTPClient(partyURL),
		guildclient.NewHTTPClient(guildURL),
		workerclient.NewHTTPClient(workerURL),
		socialclient.NewHTTPClient(socialURL),
		bootstrapReader,
		redisRuntimeReader,
	), cleanup, nil
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
