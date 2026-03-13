package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/pkg/logging"
	"github.com/xyun1996/social_backend/services/presence/internal/handler"
	redisrepo "github.com/xyun1996/social_backend/services/presence/internal/repo/redis"
	"github.com/xyun1996/social_backend/services/presence/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("presence", ":8087")
	logger := logging.New(cfg.Name, cfg.Env)
	presenceService, cleanup, err := buildPresenceService()
	if err != nil {
		logger.Error("failed to initialize presence service", "error", err)
		panic(err)
	}
	defer cleanup()
	mux := handler.NewHTTPHandler(presenceService).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildPresenceService() (*service.PresenceService, func(), error) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("PRESENCE_STORE")), "redis") {
		return service.NewPresenceService(), func() {}, nil
	}

	redisConfig := db.LoadRedisConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Username: redisConfig.Username,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, func() {}, err
	}

	store := redisrepo.NewStore(redisConfig, client)
	return service.NewPresenceServiceWithStore(store), func() {
		_ = client.Close()
	}, nil
}
