package main

import (
	"context"
	"os"

	"github.com/xyun1996/social_backend/pkg/app"
	"github.com/xyun1996/social_backend/pkg/config"
	"github.com/xyun1996/social_backend/pkg/logging"
	chatclient "github.com/xyun1996/social_backend/services/worker/internal/client/chat"
	inviteclient "github.com/xyun1996/social_backend/services/worker/internal/client/invite"
	"github.com/xyun1996/social_backend/services/worker/internal/handler"
	"github.com/xyun1996/social_backend/services/worker/internal/jobs"
	"github.com/xyun1996/social_backend/services/worker/internal/service"
)

func main() {
	cfg := config.LoadServiceConfig("worker", ":8089")
	logger := logging.New(cfg.Name, cfg.Env)
	worker := service.NewWorkerService()
	inviteURL := os.Getenv("INVITE_BASE_URL")
	chatURL := os.Getenv("CHAT_BASE_URL")
	if inviteURL != "" {
		inviteJobs := jobs.NewInviteExpireHandler(inviteclient.NewHTTPClient(inviteURL))
		worker.RegisterHandler("invite.expire", inviteJobs.Handle)
	}
	if chatURL != "" {
		chatJobs := jobs.NewChatOfflineDeliveryHandler(chatclient.NewHTTPClient(chatURL))
		worker.RegisterHandler("chat.offline_delivery", chatJobs.Handle)
	}

	mux := handler.NewHTTPHandler(worker).Routes()

	service := app.NewHTTPService(cfg.Name, cfg.Addr, logger, mux)
	if err := service.Run(context.Background()); err != nil {
		logger.Error("service exited with error", "error", err)
		panic(err)
	}
}
