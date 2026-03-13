package main

import (
	"context"
	"os"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	presenceclient "github.com/xyun1996/social_backend/services/chat/internal/client/presence"
	workerclient "github.com/xyun1996/social_backend/services/chat/internal/client/worker"
	"github.com/xyun1996/social_backend/services/chat/internal/handler"
	"github.com/xyun1996/social_backend/services/chat/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("chat", ":8084")
	logger := logging.New(cfg.Name, cfg.Env)
	presenceBaseURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://127.0.0.1:8087")
	workerBaseURL := os.Getenv("WORKER_BASE_URL")

	var scheduler service.JobScheduler
	if workerBaseURL != "" {
		scheduler = workerclient.NewHTTPClient(workerBaseURL)
	}

	mux := handler.NewHTTPHandler(service.NewChatService(
		presenceclient.NewHTTPClient(presenceBaseURL),
		scheduler,
	)).Routes()

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
