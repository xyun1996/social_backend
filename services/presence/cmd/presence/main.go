package main

import (
	"context"
	"os"
	"strings"

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
	presenceService, cleanup := buildPresenceService()
	defer cleanup()
	mux := handler.NewHTTPHandler(presenceService).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildPresenceService() (*service.PresenceService, func()) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("PRESENCE_STORE")), "redis") {
		return service.NewPresenceService(), func() {}
	}

	redisConfig := db.LoadRedisConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Username: redisConfig.Username,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	store := redisrepo.NewStore(redisConfig, client)
	return service.NewPresenceServiceWithStore(store), func() {
		_ = client.Close()
	}
}
