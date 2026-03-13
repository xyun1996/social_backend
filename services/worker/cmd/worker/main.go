package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/db"
	"github.com/xyun1996/social_backend/pkg/logging"
	chatclient "github.com/xyun1996/social_backend/services/worker/internal/client/chat"
	inviteclient "github.com/xyun1996/social_backend/services/worker/internal/client/invite"
	"github.com/xyun1996/social_backend/services/worker/internal/handler"
	"github.com/xyun1996/social_backend/services/worker/internal/jobs"
	redisrepo "github.com/xyun1996/social_backend/services/worker/internal/repo/redis"
	"github.com/xyun1996/social_backend/services/worker/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("worker", ":8089")
	logger := logging.New(cfg.Name, cfg.Env)
	worker, cleanup, err := buildWorkerService()
	if err != nil {
		logger.Error("failed to initialize worker service", "error", err)
		panic(err)
	}
	defer cleanup()

	inviteURL := os.Getenv("INVITE_BASE_URL")
	chatURL := os.Getenv("CHAT_BASE_URL")
	if inviteURL != "" {
		inviteJobs := jobs.NewInviteExpireHandler(inviteclient.NewHTTPClient(inviteURL))
		worker.RegisterHandler("invite.expire", inviteJobs.Handle)
	}
	if chatURL != "" {
		chatJobs := jobs.NewChatOfflineDeliveryHandler(chatclient.NewHTTPClient(chatURL))
		worker.RegisterHandler("chat.offline_delivery", chatJobs.Handle)
	}

	backgroundEnabled := os.Getenv("WORKER_AUTO_RUN") == "true"
	if backgroundEnabled {
		interval := intervalFromEnv(os.Getenv("WORKER_AUTO_RUN_INTERVAL_MS"), 250*time.Millisecond)
		go func() {
			_ = worker.RunBackground(context.Background(), service.BackgroundRunConfig{
				WorkerID: "worker-bg",
				Interval: interval,
				Limit:    100,
			})
		}()
	}

	mux := handler.NewHTTPHandler(worker).Routes()

	httpService := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := httpService.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func buildWorkerService() (*service.WorkerService, func(), error) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("WORKER_STORE")), "redis") {
		return service.NewWorkerService(), func() {}, nil
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

	repo := redisrepo.NewRepository(redisConfig, client)
	return service.NewWorkerServiceWithStore(repo), func() {
		_ = client.Close()
	}, nil
}

func intervalFromEnv(raw string, fallback time.Duration) time.Duration {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return time.Duration(value) * time.Millisecond
}
