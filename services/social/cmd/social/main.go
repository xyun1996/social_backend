package main

import (
	"context"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	"github.com/xyun1996/social_backend/services/social/internal/handler"
	"github.com/xyun1996/social_backend/services/social/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("social", ":8082")
	logger := logging.New(cfg.Name, cfg.Env)
	mux := handler.NewHTTPHandler(service.NewSocialService()).Routes()

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}
