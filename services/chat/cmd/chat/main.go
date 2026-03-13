package main

import (
	"context"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	"github.com/xyun1996/social_backend/services/chat/internal/handler"
	"github.com/xyun1996/social_backend/services/chat/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("chat", ":8084")
	logger := logging.New(cfg.Name, cfg.Env)
	mux := handler.NewHTTPHandler(service.NewChatService()).Routes()

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}
