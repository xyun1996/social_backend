package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/pkg/logging"
	chatclient "github.com/xyun1996/social_backend/services/gateway/internal/client/chat"
	identityclient "github.com/xyun1996/social_backend/services/gateway/internal/client/identity"
	presenceclient "github.com/xyun1996/social_backend/services/gateway/internal/client/presence"
	"github.com/xyun1996/social_backend/services/gateway/internal/handler"
	redisrepo "github.com/xyun1996/social_backend/services/gateway/internal/repo/redis"
	"github.com/xyun1996/social_backend/services/gateway/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("gateway", ":8080")
	logger := logging.New(cfg.Name, cfg.Env)
	identityBaseURL := valueOrDefault(os.Getenv("IDENTITY_BASE_URL"), "http://127.0.0.1:8081")
	presenceBaseURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://127.0.0.1:8087")
	chatBaseURL := valueOrDefault(os.Getenv("CHAT_BASE_URL"), "http://127.0.0.1:8084")
	introspector := identityclient.NewHTTPClient(identityBaseURL)
	reporter := presenceclient.NewHTTPClient(presenceBaseURL)
	chat := chatclient.NewHTTPClient(chatBaseURL)
	realtime, cleanup, err := buildRealtimeService(introspector, reporter)
	if err != nil {
		logger.Error("failed to initialize gateway realtime service", "error", err)
		panic(err)
	}
	defer cleanup()

	mux := handler.NewHTTPHandlerWithRealtime(introspector, reporter, chat, realtime).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildRealtimeService(introspector service.Introspector, reporter service.PresenceReporter) (*service.RealtimeService, func(), error) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("GATEWAY_STORE")), "redis") {
		return service.NewRealtimeService(introspector, reporter), func() {}, nil
	}

	redisConfig := db.LoadRedisConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := db.OpenRedis(ctx, redisConfig)
	if err != nil {
		return nil, func() {}, err
	}

	repo := redisrepo.NewRepository(redisConfig, client)
	return service.NewRealtimeServiceWithStores(repo, repo, introspector, reporter), func() {
		_ = client.Close()
	}, nil
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
