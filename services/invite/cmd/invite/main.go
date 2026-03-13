package main

import (
	"context"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	"github.com/xyun1996/social_backend/services/invite/internal/handler"
	"github.com/xyun1996/social_backend/services/invite/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("invite", ":8083")
	logger := logging.New(cfg.Name, cfg.Env)
	mux := handler.NewHTTPHandler(service.NewInviteService()).Routes()

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}
