package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	chatclient "github.com/xyun1996/social_backend/services/worker/internal/client/chat"
	inviteclient "github.com/xyun1996/social_backend/services/worker/internal/client/invite"
	"github.com/xyun1996/social_backend/services/worker/internal/handler"
	"github.com/xyun1996/social_backend/services/worker/internal/jobs"
	"github.com/xyun1996/social_backend/services/worker/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("worker", ":8089")
	logger := logging.New(cfg.Name, cfg.Env)
	worker := service.NewWorkerService()
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

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
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
