package main

import (
	"context"
	"os"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	guildclient "github.com/xyun1996/social_backend/services/ops/internal/client/guild"
	partyclient "github.com/xyun1996/social_backend/services/ops/internal/client/party"
	presenceclient "github.com/xyun1996/social_backend/services/ops/internal/client/presence"
	workerclient "github.com/xyun1996/social_backend/services/ops/internal/client/worker"
	"github.com/xyun1996/social_backend/services/ops/internal/handler"
	"github.com/xyun1996/social_backend/services/ops/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("ops", ":8088")
	logger := logging.New(cfg.Name, cfg.Env)
	presenceURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://localhost:8087")
	partyURL := valueOrDefault(os.Getenv("PARTY_BASE_URL"), "http://localhost:8085")
	guildURL := valueOrDefault(os.Getenv("GUILD_BASE_URL"), "http://localhost:8086")
	workerURL := valueOrDefault(os.Getenv("WORKER_BASE_URL"), "http://localhost:8089")

	mux := handler.NewHTTPHandler(service.NewOpsService(
		presenceclient.NewHTTPClient(presenceURL),
		partyclient.NewHTTPClient(partyURL),
		guildclient.NewHTTPClient(guildURL),
		workerclient.NewHTTPClient(workerURL),
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
