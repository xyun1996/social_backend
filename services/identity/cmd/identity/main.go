package main

import (
	"context"
	"net/http"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	"github.com/xyun1996/social_backend/pkg/transport"
)

func main() {
	cfg := config.LoadServiceConfig("identity", ":8081")
	logger := logging.New(cfg.Name, cfg.Env)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
			Service: cfg.Name,
			Status:  "ok",
		})
	})

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}
