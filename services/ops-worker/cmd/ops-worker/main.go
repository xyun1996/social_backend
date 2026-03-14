package main

import (
	"context"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	"github.com/xyun1996/social_backend/services/ops-worker/internal/bootstrap"
)

func main() {
	cfg := config.LoadServiceConfig("ops-worker", ":8092")
	logger := logging.New(cfg.Name, cfg.Env)

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, bootstrap.NewRouter())
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}
