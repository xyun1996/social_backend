package main

import (
	"context"
	"os"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	workerclient "github.com/xyun1996/social_backend/services/invite/internal/client/worker"
	"github.com/xyun1996/social_backend/services/invite/internal/handler"
	"github.com/xyun1996/social_backend/services/invite/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("invite", ":8083")
	logger := logging.New(cfg.Name, cfg.Env)
	workerURL := os.Getenv("WORKER_BASE_URL")

	var scheduler service.JobScheduler
	if workerURL != "" {
		scheduler = workerclient.NewHTTPClient(workerURL)
	}

	mux := handler.NewHTTPHandler(service.NewInviteService(scheduler)).Routes()

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}
