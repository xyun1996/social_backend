package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	return &HTTPService{
		name: name,
		server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
	}
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
