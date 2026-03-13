package main

import (
	"context"
	"os"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	identityclient "github.com/xyun1996/social_backend/services/gateway/internal/client/identity"
	presenceclient "github.com/xyun1996/social_backend/services/gateway/internal/client/presence"
	"github.com/xyun1996/social_backend/services/gateway/internal/handler"
)

func main() {
	cfg := config.LoadServiceConfig("gateway", ":8080")
	logger := logging.New(cfg.Name, cfg.Env)
	identityBaseURL := valueOrDefault(os.Getenv("IDENTITY_BASE_URL"), "http://127.0.0.1:8081")
	presenceBaseURL := valueOrDefault(os.Getenv("PRESENCE_BASE_URL"), "http://127.0.0.1:8087")
	mux := handler.NewHTTPHandler(
		identityclient.NewHTTPClient(identityBaseURL),
		presenceclient.NewHTTPClient(presenceBaseURL),
	).Routes()

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
