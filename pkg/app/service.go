package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xyun1996/social_backend/pkg/metrics"
	"github.com/xyun1996/social_backend/pkg/middleware"
	"github.com/xyun1996/social_backend/pkg/transport"
)

const shutdownTimeout = 10 * time.Second

// HTTPService represents a minimal long-running process with shared lifecycle handling.
type HTTPService struct {
	name   string
	server *http.Server
	logger *slog.Logger
}

// NewHTTPService builds a service wrapper around an HTTP server.
func NewHTTPService(name string, addr string, logger *slog.Logger, handler http.Handler) *HTTPService {
	registry := metrics.NewRegistry(name)
	root := http.NewServeMux()
	root.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{Service: name, Status: "ready"})
	})
	root.Handle("/metrics", registry.Handler())
	root.Handle("/", wrapHandler(name, logger, registry, handler))

	return &HTTPService{
		name: name,
		server: &http.Server{
			Addr:              addr,
			Handler:           root,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
	}
}

func wrapHandler(name string, logger *slog.Logger, registry *metrics.Registry, handler http.Handler) http.Handler {
	prefix := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	internalToken := strings.TrimSpace(os.Getenv("APP_INTERNAL_TOKEN"))
	opsToken := strings.TrimSpace(os.Getenv("OPS_API_TOKEN"))
	rateLimitRPS := envInt(prefix+"_HTTP_RATE_LIMIT_RPS", envInt("APP_HTTP_RATE_LIMIT_RPS", 0))
	rateLimitBurst := envInt(prefix+"_HTTP_RATE_LIMIT_BURST", envInt("APP_HTTP_RATE_LIMIT_BURST", rateLimitRPS))

	wrapped := handler
	wrapped = middleware.RequireInternalToken(internalToken)(wrapped)
	wrapped = middleware.RequireOpsToken(opsToken)(wrapped)
	wrapped = middleware.RateLimit(rateLimitRPS, rateLimitBurst)(wrapped)
	wrapped = middleware.AuditLog(logger)(wrapped)
	wrapped = middleware.AccessLog(logger, registry)(wrapped)
	wrapped = middleware.Recover(logger)(wrapped)
	wrapped = middleware.WithRequestContext(logger)(wrapped)
	return wrapped
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

// Run starts the service and blocks until the process receives a shutdown signal.
func (s *HTTPService) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("service starting", slog.String("addr", s.server.Addr))

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
			return
		}

		errCh <- nil
	}()

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-signalCtx.Done():
		s.logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown service %s: %w", s.name, err)
	}

	s.logger.Info("service stopped")
	return nil
}
