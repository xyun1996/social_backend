package main

import (
	"context"
	"os"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	inviteclient "github.com/xyun1996/social_backend/services/party/internal/client/invite"
	presenceclient "github.com/xyun1996/social_backend/services/party/internal/client/presence"
	"github.com/xyun1996/social_backend/services/party/internal/handler"
	"github.com/xyun1996/social_backend/services/party/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("party", ":8085")
	logger := logging.New(cfg.Name, cfg.Env)
	inviteURL := valueOrDefault(os.Getenv("INVITE_SERVICE_URL"), "http://localhost:8083")
	presenceURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://localhost:8087")
	mux := handler.NewHTTPHandler(service.NewPartyService(
		inviteclient.NewHTTPClient(inviteURL),
		presenceclient.NewHTTPClient(presenceURL),
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
